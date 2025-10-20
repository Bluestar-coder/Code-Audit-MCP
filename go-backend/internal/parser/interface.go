package parser

import (
	"context"
	"time"
)

// Language represents supported programming languages
type Language string

const (
	LanguageGo         Language = "go"
	LanguagePython     Language = "python"
	LanguageJavaScript Language = "javascript"
	LanguageJava       Language = "java"
	LanguagePHP        Language = "php"
	LanguageC          Language = "c"
	LanguageCPP        Language = "cpp"
)

// ASTNode represents a generic AST node
type ASTNode struct {
	Type       string                 `json:"type"`
	Value      string                 `json:"value,omitempty"`
	StartPos   Position               `json:"start_pos"`
	EndPos     Position               `json:"end_pos"`
	Children   []*ASTNode             `json:"children,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// Position represents a position in source code
type Position struct {
	Line   int `json:"line"`
	Column int `json:"column"`
	Offset int `json:"offset"`
}

// ParseResult contains the result of parsing
type ParseResult struct {
	AST         *ASTNode      `json:"ast"`
	Errors      []ParseError  `json:"errors,omitempty"`
	Metadata    ParseMetadata `json:"metadata"`
	Language    Language      `json:"language"`
	FilePath    string        `json:"file_path"`
	Success     bool          `json:"success"`
}

// ParseError represents a parsing error
type ParseError struct {
	Message  string   `json:"message"`
	Position Position `json:"position"`
	Severity string   `json:"severity"` // "error", "warning", "info"
	Code     string   `json:"code,omitempty"`
}

// ParseMetadata contains metadata about the parsing process
type ParseMetadata struct {
	ParseTime       time.Duration `json:"parse_time"`
	TotalLines      int           `json:"total_lines"`
	TotalNodes      int           `json:"total_nodes"`
	TotalFunctions  int           `json:"total_functions"`
	TotalClasses    int           `json:"total_classes"`
	TotalVariables  int           `json:"total_variables"`
	LanguageVersion string        `json:"language_version"`
	ParserVersion   string        `json:"parser_version"`
}

// ParseOptions contains options for parsing
type ParseOptions struct {
	IncludeComments  bool `json:"include_comments"`
	IncludeMetadata  bool `json:"include_metadata"`
	MaxDepth         int  `json:"max_depth"`
	TimeoutSeconds   int  `json:"timeout_seconds"`
	StrictMode       bool `json:"strict_mode"`
}

// Parser interface defines the contract for language-specific parsers
type Parser interface {
	// Parse parses source code and returns AST
	Parse(ctx context.Context, content []byte, filePath string, options ParseOptions) (*ParseResult, error)
	
	// GetLanguage returns the language this parser supports
	GetLanguage() Language
	
	// GetVersion returns the parser version
	GetVersion() string
	
	// IsSupported checks if the file extension is supported
	IsSupported(filePath string) bool
	
	// Validate validates the parser configuration
	Validate() error
}

// ParserRegistry manages multiple language parsers
type ParserRegistry struct {
	parsers map[Language]Parser
}

// NewParserRegistry creates a new parser registry
func NewParserRegistry() *ParserRegistry {
	return &ParserRegistry{
		parsers: make(map[Language]Parser),
	}
}

// Register registers a parser for a language
func (r *ParserRegistry) Register(parser Parser) error {
	if err := parser.Validate(); err != nil {
		return err
	}
	r.parsers[parser.GetLanguage()] = parser
	return nil
}

// GetParser returns a parser for the given language
func (r *ParserRegistry) GetParser(language Language) (Parser, bool) {
	parser, exists := r.parsers[language]
	return parser, exists
}

// GetSupportedLanguages returns all supported languages
func (r *ParserRegistry) GetSupportedLanguages() []Language {
	languages := make([]Language, 0, len(r.parsers))
	for lang := range r.parsers {
		languages = append(languages, lang)
	}
	return languages
}

// DetectLanguage attempts to detect the language from file path
func DetectLanguage(filePath string) Language {
	// Simple file extension based detection
	switch {
	case hasExtension(filePath, ".go"):
		return LanguageGo
	case hasExtension(filePath, ".py", ".pyw"):
		return LanguagePython
	case hasExtension(filePath, ".js", ".jsx", ".ts", ".tsx"):
		return LanguageJavaScript
	case hasExtension(filePath, ".java"):
		return LanguageJava
	case hasExtension(filePath, ".php"):
		return LanguagePHP
	case hasExtension(filePath, ".c", ".h"):
		return LanguageC
	case hasExtension(filePath, ".cpp", ".cxx", ".cc", ".hpp"):
		return LanguageCPP
	default:
		return ""
	}
}

// hasExtension checks if file has any of the given extensions
func hasExtension(filePath string, extensions ...string) bool {
	for _, ext := range extensions {
		if len(filePath) >= len(ext) && filePath[len(filePath)-len(ext):] == ext {
			return true
		}
	}
	return false
}