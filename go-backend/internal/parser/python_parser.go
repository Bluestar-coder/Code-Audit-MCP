package parser

import (
	"context"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// PythonParser implements Parser interface for Python language
type PythonParser struct {
	version string
}

// NewPythonParser creates a new Python parser
func NewPythonParser() *PythonParser {
	return &PythonParser{
		version: "1.0.0",
	}
}

// Parse implements Parser interface
func (p *PythonParser) Parse(ctx context.Context, content []byte, filePath string, options ParseOptions) (*ParseResult, error) {
	startTime := time.Now()
	
	source := string(content)
	lines := strings.Split(source, "\n")
	
	// Create root AST node
	astNode := &ASTNode{
		Type: "Module",
		StartPos: Position{Line: 1, Column: 1, Offset: 0},
		EndPos:   Position{Line: len(lines), Column: len(lines[len(lines)-1]), Offset: len(source)},
		Attributes: map[string]interface{}{
			"filename": filePath,
		},
	}
	
	// Parse the Python source using regex patterns (simplified approach)
	children, metadata := p.parseStatements(lines)
	astNode.Children = children
	
	// Update metadata
	metadata.ParseTime = time.Since(startTime)
	metadata.TotalLines = len(lines)
	metadata.ParserVersion = p.version
	metadata.LanguageVersion = "python3"
	
	return &ParseResult{
		AST:      astNode,
		Success:  true,
		Language: LanguagePython,
		FilePath: filePath,
		Metadata: metadata,
	}, nil
}

// parseStatements parses Python statements using regex patterns
func (p *PythonParser) parseStatements(lines []string) ([]*ASTNode, ParseMetadata) {
	var children []*ASTNode
	metadata := ParseMetadata{}
	
	// Regex patterns for Python constructs
	funcPattern := regexp.MustCompile(`^\s*(def|async\s+def)\s+(\w+)\s*\(([^)]*)\)\s*:`)
	classPattern := regexp.MustCompile(`^\s*class\s+(\w+)(\([^)]*\))?\s*:`)
	importPattern := regexp.MustCompile(`^\s*(import|from)\s+(.+)`)
	assignPattern := regexp.MustCompile(`^\s*(\w+)\s*=`)
	commentPattern := regexp.MustCompile(`^\s*#`)
	
	for lineNum, line := range lines {
		lineNum++ // 1-based line numbers
		
		// Skip empty lines and comments
		if strings.TrimSpace(line) == "" || commentPattern.MatchString(line) {
			continue
		}
		
		metadata.TotalNodes++
		
		// Function definitions
		if matches := funcPattern.FindStringSubmatch(line); matches != nil {
			funcNode := &ASTNode{
				Type:  "FunctionDef",
				Value: matches[2],
				StartPos: Position{
					Line:   lineNum,
					Column: 1,
					Offset: p.calculateOffset(lines, lineNum-1),
				},
				EndPos: Position{
					Line:   lineNum,
					Column: len(line),
					Offset: p.calculateOffset(lines, lineNum-1) + len(line),
				},
				Attributes: map[string]interface{}{
					"name":   matches[2],
					"async":  strings.Contains(matches[1], "async"),
					"params": strings.TrimSpace(matches[3]),
				},
			}
			children = append(children, funcNode)
			metadata.TotalFunctions++
			continue
		}
		
		// Class definitions
		if matches := classPattern.FindStringSubmatch(line); matches != nil {
			classNode := &ASTNode{
				Type:  "ClassDef",
				Value: matches[1],
				StartPos: Position{
					Line:   lineNum,
					Column: 1,
					Offset: p.calculateOffset(lines, lineNum-1),
				},
				EndPos: Position{
					Line:   lineNum,
					Column: len(line),
					Offset: p.calculateOffset(lines, lineNum-1) + len(line),
				},
				Attributes: map[string]interface{}{
					"name":    matches[1],
					"bases":   strings.TrimSpace(matches[2]),
				},
			}
			children = append(children, classNode)
			metadata.TotalClasses++
			continue
		}
		
		// Import statements
		if matches := importPattern.FindStringSubmatch(line); matches != nil {
			importNode := &ASTNode{
				Type:  "Import",
				Value: matches[2],
				StartPos: Position{
					Line:   lineNum,
					Column: 1,
					Offset: p.calculateOffset(lines, lineNum-1),
				},
				EndPos: Position{
					Line:   lineNum,
					Column: len(line),
					Offset: p.calculateOffset(lines, lineNum-1) + len(line),
				},
				Attributes: map[string]interface{}{
					"type":   matches[1],
					"module": strings.TrimSpace(matches[2]),
				},
			}
			children = append(children, importNode)
			continue
		}
		
		// Variable assignments
		if matches := assignPattern.FindStringSubmatch(line); matches != nil {
			assignNode := &ASTNode{
				Type:  "Assign",
				Value: matches[1],
				StartPos: Position{
					Line:   lineNum,
					Column: 1,
					Offset: p.calculateOffset(lines, lineNum-1),
				},
				EndPos: Position{
					Line:   lineNum,
					Column: len(line),
					Offset: p.calculateOffset(lines, lineNum-1) + len(line),
				},
				Attributes: map[string]interface{}{
					"target": matches[1],
					"value":  strings.TrimSpace(line[strings.Index(line, "=")+1:]),
				},
			}
			children = append(children, assignNode)
			metadata.TotalVariables++
			continue
		}
		
		// Generic statement
		stmtNode := &ASTNode{
			Type:  "Statement",
			Value: strings.TrimSpace(line),
			StartPos: Position{
				Line:   lineNum,
				Column: 1,
				Offset: p.calculateOffset(lines, lineNum-1),
			},
			EndPos: Position{
				Line:   lineNum,
				Column: len(line),
				Offset: p.calculateOffset(lines, lineNum-1) + len(line),
			},
			Attributes: map[string]interface{}{
				"content": strings.TrimSpace(line),
			},
		}
		children = append(children, stmtNode)
	}
	
	return children, metadata
}

// calculateOffset calculates the byte offset for a given line
func (p *PythonParser) calculateOffset(lines []string, lineIndex int) int {
	offset := 0
	for i := 0; i < lineIndex && i < len(lines); i++ {
		offset += len(lines[i]) + 1 // +1 for newline
	}
	return offset
}

// GetLanguage returns the supported language
func (p *PythonParser) GetLanguage() Language {
	return LanguagePython
}

// GetVersion returns the parser version
func (p *PythonParser) GetVersion() string {
	return p.version
}

// IsSupported checks if the file is a Python file
func (p *PythonParser) IsSupported(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return ext == ".py" || ext == ".pyw"
}

// Validate validates the parser configuration
func (p *PythonParser) Validate() error {
	// Basic Python parser doesn't need external dependencies
	return nil
}