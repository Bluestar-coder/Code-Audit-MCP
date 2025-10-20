package grpc

import (
	"context"
	"fmt"
	"log"

	pb "code-audit-mcp/proto"
	"code-audit-mcp/internal/parser"
)

// ASTParserService implements pb.ASTParserServer
type ASTParserService struct {
	pb.UnimplementedASTParserServer
	manager *parser.Manager
}

// NewASTParserService creates a new AST parser service
func NewASTParserService() *ASTParserService {
	return &ASTParserService{
		manager: parser.NewManager(),
	}
}

// ParseFile implements pb.ASTParser/ParseFile
func (s *ASTParserService) ParseFile(ctx context.Context, req *pb.ParseRequest) (*pb.ParseResponse, error) {
	log.Printf("📄 Parsing file: %s (language: %s)", req.FilePath, req.Language)

	// Convert language string to parser.Language
	language := parser.Language(req.Language)
	if language == "" {
		language = parser.DetectLanguage(req.FilePath)
	}

	// Set up parse options
	options := parser.ParseOptions{
		IncludeComments: req.IncludeMetadata,
		IncludeMetadata: req.IncludeMetadata,
		TimeoutSeconds:  30,
		MaxDepth:        100,
		StrictMode:      false,
	}

	// Parse the file
	result, err := s.manager.ParseFile(ctx, req.Content, req.FilePath, language, options)
	if err != nil {
		log.Printf("❌ Parsing failed: %v", err)
		return &pb.ParseResponse{
			FilePath: req.FilePath,
			Success:  false,
			Errors:   []*pb.ParseError{{Message: err.Error()}},
		}, nil
	}

	// Serialize AST to JSON
	astData, err := s.manager.SerializeAST(result.AST)
	if err != nil {
		log.Printf("❌ AST serialization failed: %v", err)
		return &pb.ParseResponse{
			FilePath: req.FilePath,
			Success:  false,
			Errors:   []*pb.ParseError{{Message: fmt.Sprintf("AST serialization failed: %v", err)}},
		}, nil
	}

	// Convert parse errors
	var errors []*pb.ParseError
	for _, parseErr := range result.Errors {
		errors = append(errors, &pb.ParseError{
			Line:      int32(parseErr.Position.Line),
			Column:    int32(parseErr.Position.Column),
			Message:   parseErr.Message,
			ErrorType: parseErr.Severity,
		})
	}

	// Create response
	resp := &pb.ParseResponse{
		FilePath: req.FilePath,
		Success:  result.Success,
		AstData:  astData,
		Errors:   errors,
		Metadata: &pb.ParseMetadata{
			ParseTimeMs:     result.Metadata.ParseTime.Milliseconds(),
			TotalLines:      int32(result.Metadata.TotalLines),
			TotalFunctions:  int32(result.Metadata.TotalFunctions),
			TotalClasses:    int32(result.Metadata.TotalClasses),
			LanguageVersion: result.Metadata.LanguageVersion,
		},
	}

	log.Printf("✅ Parsing completed: %s (nodes: %d, functions: %d, classes: %d)", 
		req.FilePath, result.Metadata.TotalNodes, result.Metadata.TotalFunctions, result.Metadata.TotalClasses)

	return resp, nil
}

// ParseBatch implements pb.ASTParser/ParseBatch
func (s *ASTParserService) ParseBatch(req *pb.BatchParseRequest, stream pb.ASTParser_ParseBatchServer) error {
	log.Printf("📦 Batch parsing %d files", len(req.Requests))

	// Convert requests to parser format
	parseRequests := make([]parser.ParseRequest, len(req.Requests))
	for i, pbReq := range req.Requests {
		parseRequests[i] = parser.ParseRequest{
			Content:  pbReq.Content,
			FilePath: pbReq.FilePath,
			Language: parser.Language(pbReq.Language),
		}
	}

	// Set up parse options
	options := parser.ParseOptions{
		IncludeComments: true,
		IncludeMetadata: true,
		TimeoutSeconds:  30,
		MaxDepth:        100,
		StrictMode:      false,
	}

	// Start batch parsing
	ctx := stream.Context()
	resultChan := s.manager.ParseBatch(ctx, parseRequests, options)

	// Stream results as they come
	count := 0
	for result := range resultChan {
		count++
		log.Printf("  [%d/%d] Completed parsing %s", count, len(req.Requests), result.FilePath)

		// Serialize AST to JSON
		var astData []byte
		var errors []*pb.ParseError
		
		if result.Success && result.AST != nil {
			var err error
			astData, err = s.manager.SerializeAST(result.AST)
			if err != nil {
				log.Printf("❌ AST serialization failed for %s: %v", result.FilePath, err)
				result.Success = false
				errors = append(errors, &pb.ParseError{
					Message: fmt.Sprintf("AST serialization failed: %v", err),
				})
			}
		}

		// Convert parse errors
		for _, parseErr := range result.Errors {
			errors = append(errors, &pb.ParseError{
				Line:      int32(parseErr.Position.Line),
				Column:    int32(parseErr.Position.Column),
				Message:   parseErr.Message,
				ErrorType: parseErr.Severity,
			})
		}

		// Create response
		resp := &pb.ParseResponse{
			FilePath: result.FilePath,
			Success:  result.Success,
			AstData:  astData,
			Errors:   errors,
			Metadata: &pb.ParseMetadata{
				ParseTimeMs:     result.Metadata.ParseTime.Milliseconds(),
				TotalLines:      int32(result.Metadata.TotalLines),
				TotalFunctions:  int32(result.Metadata.TotalFunctions),
				TotalClasses:    int32(result.Metadata.TotalClasses),
				LanguageVersion: result.Metadata.LanguageVersion,
			},
		}

		// Send response
		if err := stream.Send(resp); err != nil {
			log.Printf("❌ Error sending response for %s: %v", result.FilePath, err)
			return err
		}
	}

	log.Printf("✅ Batch parsing completed (%d files processed)", count)
	return nil
}
