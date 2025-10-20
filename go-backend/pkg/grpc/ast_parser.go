package grpc

import (
	"context"
	"fmt"
	"log"

	pb "code-audit-mcp/proto"
)

// ASTParserService implements pb.ASTParserServer
type ASTParserService struct {
	pb.UnimplementedASTParserServer
}

// NewASTParserService creates a new AST parser service
func NewASTParserService() *ASTParserService {
	return &ASTParserService{}
}

// ParseFile implements pb.ASTParser/ParseFile
func (s *ASTParserService) ParseFile(ctx context.Context, req *pb.ParseRequest) (*pb.ParseResponse, error) {
	log.Printf("📄 Parsing file: %s (language: %s)", req.FilePath, req.Language)

	// TODO: Implement actual AST parsing using go-tree-sitter
	// For now, return a placeholder response

	resp := &pb.ParseResponse{
		FilePath: req.FilePath,
		Success:  true,
		AstData:  []byte(`{"type": "placeholder", "status": "parsing not yet implemented"}`),
		Errors:   nil,
		Metadata: &pb.ParseMetadata{
			ParseTimeMs:     100,
			TotalLines:      0,
			TotalFunctions:  0,
			TotalClasses:    0,
			LanguageVersion: "1.0",
		},
	}

	return resp, nil
}

// ParseBatch implements pb.ASTParser/ParseBatch
func (s *ASTParserService) ParseBatch(req *pb.BatchParseRequest, stream pb.ASTParser_ParseBatchServer) error {
	log.Printf("📦 Batch parsing %d files", len(req.Requests))

	// TODO: Implement concurrent batch parsing
	// For now, process requests sequentially

	for i, parseReq := range req.Requests {
		log.Printf("  [%d/%d] Parsing %s", i+1, len(req.Requests), parseReq.FilePath)

		resp := &pb.ParseResponse{
			FilePath: parseReq.FilePath,
			Success:  true,
			AstData:  []byte(fmt.Sprintf(`{"index": %d, "status": "batch parsing"}`, i)),
			Metadata: &pb.ParseMetadata{
				ParseTimeMs: 50,
			},
		}

		if err := stream.Send(resp); err != nil {
			log.Printf("❌ Error sending response: %v", err)
			return err
		}
	}

	log.Printf("✅ Batch parsing completed")
	return nil
}
