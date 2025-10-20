package parser

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"time"
)

// GoParser implements Parser interface for Go language
type GoParser struct {
	version string
}

// NewGoParser creates a new Go parser
func NewGoParser() *GoParser {
	return &GoParser{
		version: "1.0.0",
	}
}

// Parse implements Parser interface
func (p *GoParser) Parse(ctx context.Context, content []byte, filePath string, options ParseOptions) (*ParseResult, error) {
	startTime := time.Now()
	
	// Create file set for position tracking
	fset := token.NewFileSet()
	
	// Parse mode configuration
	mode := parser.ParseComments
	if options.StrictMode {
		mode |= parser.DeclarationErrors
	}
	
	// Parse the Go source code
	file, err := parser.ParseFile(fset, filePath, content, mode)
	if err != nil {
		return &ParseResult{
			Success:  false,
			Language: LanguageGo,
			FilePath: filePath,
			Errors:   []ParseError{{Message: err.Error(), Severity: "error"}},
			Metadata: ParseMetadata{
				ParseTime:     time.Since(startTime),
				ParserVersion: p.version,
			},
		}, nil
	}
	
	// Convert Go AST to our generic AST format
	astNode := p.convertASTNode(file, fset)
	
	// Calculate metadata
	metadata := p.calculateMetadata(file, fset, startTime)
	
	return &ParseResult{
		AST:      astNode,
		Success:  true,
		Language: LanguageGo,
		FilePath: filePath,
		Metadata: metadata,
	}, nil
}

// convertASTNode converts Go ast.Node to our generic ASTNode
func (p *GoParser) convertASTNode(node ast.Node, fset *token.FileSet) *ASTNode {
	if node == nil {
		return nil
	}
	
	pos := fset.Position(node.Pos())
	end := fset.Position(node.End())
	
	astNode := &ASTNode{
		Type: fmt.Sprintf("%T", node),
		StartPos: Position{
			Line:   pos.Line,
			Column: pos.Column,
			Offset: pos.Offset,
		},
		EndPos: Position{
			Line:   end.Line,
			Column: end.Column,
			Offset: end.Offset,
		},
		Attributes: make(map[string]interface{}),
	}
	
	// Remove package prefix from type name
	if strings.Contains(astNode.Type, ".") {
		parts := strings.Split(astNode.Type, ".")
		astNode.Type = parts[len(parts)-1]
	}
	
	// Handle specific node types
	switch n := node.(type) {
	case *ast.File:
		astNode.Attributes["package"] = n.Name.Name
		astNode.Children = p.convertChildren(n, fset)
		
	case *ast.FuncDecl:
		if n.Name != nil {
			astNode.Value = n.Name.Name
			astNode.Attributes["name"] = n.Name.Name
		}
		if n.Recv != nil {
			astNode.Attributes["receiver"] = true
		}
		astNode.Children = p.convertChildren(n, fset)
		
	case *ast.GenDecl:
		astNode.Attributes["token"] = n.Tok.String()
		astNode.Children = p.convertChildren(n, fset)
		
	case *ast.TypeSpec:
		if n.Name != nil {
			astNode.Value = n.Name.Name
			astNode.Attributes["name"] = n.Name.Name
		}
		astNode.Children = p.convertChildren(n, fset)
		
	case *ast.ValueSpec:
		if len(n.Names) > 0 {
			names := make([]string, len(n.Names))
			for i, name := range n.Names {
				names[i] = name.Name
			}
			astNode.Attributes["names"] = names
			astNode.Value = strings.Join(names, ", ")
		}
		astNode.Children = p.convertChildren(n, fset)
		
	case *ast.Ident:
		astNode.Value = n.Name
		astNode.Attributes["name"] = n.Name
		
	case *ast.BasicLit:
		astNode.Value = n.Value
		astNode.Attributes["kind"] = n.Kind.String()
		
	case *ast.CallExpr:
		astNode.Children = p.convertChildren(n, fset)
		
	case *ast.BlockStmt:
		astNode.Children = p.convertChildren(n, fset)
		
	default:
		astNode.Children = p.convertChildren(n, fset)
	}
	
	return astNode
}

// convertChildren converts child nodes
func (p *GoParser) convertChildren(node ast.Node, fset *token.FileSet) []*ASTNode {
	var children []*ASTNode
	
	ast.Inspect(node, func(n ast.Node) bool {
		if n == nil || n == node {
			return true
		}
		
		// Only include direct children
		parent := p.findParent(node, n)
		if parent == node {
			child := p.convertASTNode(n, fset)
			if child != nil {
				children = append(children, child)
			}
			return false // Don't traverse deeper
		}
		
		return true
	})
	
	return children
}

// findParent finds the parent of a node (simplified implementation)
func (p *GoParser) findParent(root, target ast.Node) ast.Node {
	var parent ast.Node
	
	ast.Inspect(root, func(n ast.Node) bool {
		if n == nil {
			return true
		}
		
		// Check if any child of n is target
		ast.Inspect(n, func(child ast.Node) bool {
			if child == target {
				parent = n
				return false
			}
			return child != n // Don't go deeper than immediate children
		})
		
		return parent == nil
	})
	
	return parent
}

// calculateMetadata calculates parsing metadata
func (p *GoParser) calculateMetadata(file *ast.File, fset *token.FileSet, startTime time.Time) ParseMetadata {
	metadata := ParseMetadata{
		ParseTime:       time.Since(startTime),
		ParserVersion:   p.version,
		LanguageVersion: "go1.21", // Could be detected from go.mod
	}
	
	// Count various elements
	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			return true
		}
		
		metadata.TotalNodes++
		
		switch n.(type) {
		case *ast.FuncDecl:
			metadata.TotalFunctions++
		case *ast.TypeSpec:
			metadata.TotalClasses++ // Structs/interfaces count as classes
		case *ast.ValueSpec:
			metadata.TotalVariables++
		}
		
		return true
	})
	
	// Calculate total lines
	if file.End().IsValid() {
		endPos := fset.Position(file.End())
		metadata.TotalLines = endPos.Line
	}
	
	return metadata
}

// GetLanguage returns the supported language
func (p *GoParser) GetLanguage() Language {
	return LanguageGo
}

// GetVersion returns the parser version
func (p *GoParser) GetVersion() string {
	return p.version
}

// IsSupported checks if the file is a Go file
func (p *GoParser) IsSupported(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return ext == ".go"
}

// Validate validates the parser configuration
func (p *GoParser) Validate() error {
	// Go parser doesn't need external dependencies
	return nil
}