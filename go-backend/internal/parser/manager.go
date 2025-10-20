package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// Manager manages all language parsers
type Manager struct {
	registry *ParserRegistry
}

// NewManager creates a new parser manager
func NewManager() *Manager {
	manager := &Manager{
		registry: NewParserRegistry(),
	}
	
	// Register built-in parsers
	manager.registerBuiltinParsers()
	
	return manager
}

// registerBuiltinParsers registers all built-in parsers
func (m *Manager) registerBuiltinParsers() {
	parsers := []Parser{
		NewGoParser(),
		NewPythonParser(),
		NewJavaScriptParser(),
		NewTypeScriptParser(),
	}
	
	for _, parser := range parsers {
		if err := m.registry.Register(parser); err != nil {
			log.Printf("❌ Failed to register %s parser: %v", parser.GetLanguage(), err)
		} else {
			log.Printf("✅ Registered %s parser (version %s)", parser.GetLanguage(), parser.GetVersion())
		}
	}
}

// ParseFile parses a single file
func (m *Manager) ParseFile(ctx context.Context, content []byte, filePath string, language Language, options ParseOptions) (*ParseResult, error) {
	// Auto-detect language if not specified
	if language == "" {
		language = DetectLanguage(filePath)
		if language == "" {
			return nil, fmt.Errorf("unable to detect language for file: %s", filePath)
		}
	}
	
	// Get parser for the language
	parser, exists := m.registry.GetParser(language)
	if !exists {
		return nil, fmt.Errorf("no parser available for language: %s", language)
	}
	
	// Set default options
	if options.TimeoutSeconds == 0 {
		options.TimeoutSeconds = 30
	}
	if options.MaxDepth == 0 {
		options.MaxDepth = 100
	}
	
	// Create context with timeout
	parseCtx, cancel := context.WithTimeout(ctx, time.Duration(options.TimeoutSeconds)*time.Second)
	defer cancel()
	
	// Parse the file
	result, err := parser.Parse(parseCtx, content, filePath, options)
	if err != nil {
		return nil, fmt.Errorf("parsing failed: %w", err)
	}
	
	return result, nil
}

// ParseBatch parses multiple files concurrently
func (m *Manager) ParseBatch(ctx context.Context, requests []ParseRequest, options ParseOptions) <-chan *ParseResult {
	resultChan := make(chan *ParseResult, len(requests))
	
	go func() {
		defer close(resultChan)
		
		// Process requests concurrently (with a reasonable limit)
		semaphore := make(chan struct{}, 10) // Limit to 10 concurrent parses
		
		for _, req := range requests {
			select {
			case <-ctx.Done():
				return
			case semaphore <- struct{}{}:
				go func(request ParseRequest) {
					defer func() { <-semaphore }()
					
					result, err := m.ParseFile(ctx, request.Content, request.FilePath, request.Language, options)
					if err != nil {
						// Create error result
						result = &ParseResult{
							Success:  false,
							Language: request.Language,
							FilePath: request.FilePath,
							Errors: []ParseError{{
								Message:  err.Error(),
								Severity: "error",
							}},
							Metadata: ParseMetadata{
								ParseTime: 0,
							},
						}
					}
					
					select {
					case resultChan <- result:
					case <-ctx.Done():
						return
					}
				}(req)
			}
		}
		
		// Wait for all goroutines to complete
		for i := 0; i < cap(semaphore); i++ {
			semaphore <- struct{}{}
		}
	}()
	
	return resultChan
}

// GetSupportedLanguages returns all supported languages
func (m *Manager) GetSupportedLanguages() []Language {
	return m.registry.GetSupportedLanguages()
}

// GetParserInfo returns information about a specific parser
func (m *Manager) GetParserInfo(language Language) (map[string]interface{}, error) {
	parser, exists := m.registry.GetParser(language)
	if !exists {
		return nil, fmt.Errorf("no parser available for language: %s", language)
	}
	
	return map[string]interface{}{
		"language": parser.GetLanguage(),
		"version":  parser.GetVersion(),
		"status":   "available",
	}, nil
}

// SerializeAST serializes AST to JSON
func (m *Manager) SerializeAST(ast *ASTNode) ([]byte, error) {
	return json.Marshal(ast)
}

// DeserializeAST deserializes AST from JSON
func (m *Manager) DeserializeAST(data []byte) (*ASTNode, error) {
	var ast ASTNode
	err := json.Unmarshal(data, &ast)
	return &ast, err
}

// ParseRequest represents a single parse request
type ParseRequest struct {
	Content  []byte   `json:"content"`
	FilePath string   `json:"file_path"`
	Language Language `json:"language"`
}

// ValidateParseOptions validates parse options
func ValidateParseOptions(options *ParseOptions) {
	if options.TimeoutSeconds <= 0 {
		options.TimeoutSeconds = 30
	}
	if options.MaxDepth <= 0 {
		options.MaxDepth = 100
	}
	if options.TimeoutSeconds > 300 {
		options.TimeoutSeconds = 300 // Max 5 minutes
	}
	if options.MaxDepth > 1000 {
		options.MaxDepth = 1000 // Reasonable limit
	}
}