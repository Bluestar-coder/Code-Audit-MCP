package parser

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"time"
)

// JavaScriptParser 实现JavaScript语言的AST解析器
type JavaScriptParser struct{}

// NewJavaScriptParser 创建新的JavaScript解析器实例
func NewJavaScriptParser() *JavaScriptParser {
	return &JavaScriptParser{}
}

// Parse 解析JavaScript源代码
func (p *JavaScriptParser) Parse(ctx context.Context, content []byte, filePath string, options ParseOptions) (*ParseResult, error) {
	startTime := time.Now()
	source := string(content)
	lines := strings.Split(source, "\n")
	
	// 创建根节点
	root := &ASTNode{
		Type:     "Program",
		Value:    "program",
		StartPos: Position{Line: 1, Column: 1},
		EndPos:   Position{Line: len(lines), Column: len(lines[len(lines)-1])},
		Children: []*ASTNode{},
	}
	
	var errors []ParseError
	var functions, classes, variables int
	
	for i, line := range lines {
		lineNum := i + 1
		trimmedLine := strings.TrimSpace(line)
		
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "//") {
			continue
		}
		
		// 解析函数声明
		if nodes := p.parseFunctionDeclaration(line, lineNum); len(nodes) > 0 {
			root.Children = append(root.Children, nodes...)
			functions += len(nodes)
		}
		
		// 解析类声明
		if nodes := p.parseClassDeclaration(line, lineNum); len(nodes) > 0 {
			root.Children = append(root.Children, nodes...)
			classes += len(nodes)
		}
		
		// 解析变量声明
		if nodes := p.parseVariableDeclaration(line, lineNum); len(nodes) > 0 {
			root.Children = append(root.Children, nodes...)
			variables += len(nodes)
		}
		
		// 解析导入语句
		if nodes := p.parseImportStatement(line, lineNum); len(nodes) > 0 {
			root.Children = append(root.Children, nodes...)
		}
		
		// 解析导出语句
		if nodes := p.parseExportStatement(line, lineNum); len(nodes) > 0 {
			root.Children = append(root.Children, nodes...)
		}
	}
	
	// 计算元数据
	metadata := ParseMetadata{
		ParseTime:       time.Since(startTime),
		TotalLines:      len(lines),
		TotalFunctions:  functions,
		TotalClasses:    classes,
		TotalVariables:  variables,
		LanguageVersion: "ES2020",
		ParserVersion:   "1.0.0",
	}
	
	return &ParseResult{
		AST:      root,
		Errors:   errors,
		Metadata: metadata,
		Language: LanguageJavaScript,
		FilePath: filePath,
		Success:  len(errors) == 0,
	}, nil
}

// parseFunctionDeclaration 解析函数声明
func (p *JavaScriptParser) parseFunctionDeclaration(line string, lineNum int) []*ASTNode {
	var nodes []*ASTNode
	
	// 匹配各种函数声明模式
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`function\s+(\w+)\s*\([^)]*\)`),                    // function name()
		regexp.MustCompile(`const\s+(\w+)\s*=\s*\([^)]*\)\s*=>`),             // const name = () =>
		regexp.MustCompile(`let\s+(\w+)\s*=\s*\([^)]*\)\s*=>`),               // let name = () =>
		regexp.MustCompile(`var\s+(\w+)\s*=\s*\([^)]*\)\s*=>`),               // var name = () =>
		regexp.MustCompile(`(\w+)\s*:\s*\([^)]*\)\s*=>`),                     // name: () =>
		regexp.MustCompile(`(\w+)\s*:\s*function\s*\([^)]*\)`),               // name: function()
		regexp.MustCompile(`async\s+function\s+(\w+)\s*\([^)]*\)`),           // async function name()
		regexp.MustCompile(`(\w+)\s*\([^)]*\)\s*\{`),                         // method() {
	}
	
	for _, pattern := range patterns {
		matches := pattern.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) > 1 {
				funcName := match[1]
				node := &ASTNode{
					Type:        "FunctionDeclaration",
					Value:       funcName,
					StartPos:    Position{Line: lineNum, Column: strings.Index(line, funcName) + 1},
					EndPos:      Position{Line: lineNum, Column: strings.Index(line, funcName) + len(funcName)},
					Children:    []*ASTNode{},
					Attributes: map[string]interface{}{
						"name":       funcName,
						"async":      strings.Contains(line, "async"),
					},
				}
				nodes = append(nodes, node)
			}
		}
	}
	
	return nodes
}

// parseClassDeclaration 解析类声明
func (p *JavaScriptParser) parseClassDeclaration(line string, lineNum int) []*ASTNode {
	var nodes []*ASTNode
	
	// 匹配类声明
	classPattern := regexp.MustCompile(`class\s+(\w+)(?:\s+extends\s+(\w+))?\s*\{?`)
	matches := classPattern.FindAllStringSubmatch(line, -1)
	
	for _, match := range matches {
		if len(match) > 1 {
			className := match[1]
			node := &ASTNode{
			Type:        "ClassDeclaration",
			Value:       className,
			StartPos:    Position{Line: lineNum, Column: strings.Index(line, className) + 1},
			EndPos:      Position{Line: lineNum, Column: strings.Index(line, className) + len(className)},
			Children:    []*ASTNode{},
			Attributes:  map[string]interface{}{
				"name": className,
			},
		}
		
		// 如果有继承，添加到属性中
		if len(match) > 2 && match[2] != "" {
			node.Attributes["extends"] = match[2]
		}
			
			nodes = append(nodes, node)
		}
	}
	
	return nodes
}

// parseVariableDeclaration 解析变量声明
func (p *JavaScriptParser) parseVariableDeclaration(line string, lineNum int) []*ASTNode {
	var nodes []*ASTNode
	
	// 匹配变量声明
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`const\s+(\w+)`),
		regexp.MustCompile(`let\s+(\w+)`),
		regexp.MustCompile(`var\s+(\w+)`),
	}
	
	for _, pattern := range patterns {
		matches := pattern.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) > 1 {
				varName := match[1]
				node := &ASTNode{
					Type:        "VariableDeclaration",
					Value:       varName,
					StartPos:    Position{Line: lineNum, Column: strings.Index(line, varName) + 1},
					EndPos:      Position{Line: lineNum, Column: strings.Index(line, varName) + len(varName)},
					Children:    []*ASTNode{},
					Attributes:  map[string]interface{}{
						"name": varName,
					},
				}
				nodes = append(nodes, node)
			}
		}
	}
	
	return nodes
}

// parseImportStatement 解析导入语句
func (p *JavaScriptParser) parseImportStatement(line string, lineNum int) []*ASTNode {
	var nodes []*ASTNode
	
	// 匹配导入语句
	importPattern := regexp.MustCompile(`import\s+(?:(\w+)|{([^}]+)}|\*\s+as\s+(\w+))\s+from\s+['"]([^'"]+)['"]`)
	matches := importPattern.FindAllStringSubmatch(line, -1)
	
	for _, match := range matches {
		if len(match) > 4 {
			var importName string
			if match[1] != "" {
				importName = match[1] // default import
			} else if match[2] != "" {
				importName = "{" + match[2] + "}" // named imports
			} else if match[3] != "" {
				importName = match[3] // namespace import
			}
			
			node := &ASTNode{
				Type:        "ImportDeclaration",
				Value:       importName,
				StartPos:    Position{Line: lineNum, Column: 1},
				EndPos:      Position{Line: lineNum, Column: len(line)},
				Children:    []*ASTNode{},
				Attributes: map[string]interface{}{
					"name":   importName,
					"source": match[4],
				},
			}
			nodes = append(nodes, node)
		}
	}
	
	return nodes
}

// parseExportStatement 解析导出语句
func (p *JavaScriptParser) parseExportStatement(line string, lineNum int) []*ASTNode {
	var nodes []*ASTNode
	
	// 匹配导出语句
	exportPatterns := []*regexp.Regexp{
		regexp.MustCompile(`export\s+default\s+(\w+)`),
		regexp.MustCompile(`export\s+{([^}]+)}`),
		regexp.MustCompile(`export\s+(?:const|let|var)\s+(\w+)`),
		regexp.MustCompile(`export\s+function\s+(\w+)`),
		regexp.MustCompile(`export\s+class\s+(\w+)`),
	}
	
	for _, pattern := range exportPatterns {
		matches := pattern.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) > 1 {
				exportName := match[1]
				node := &ASTNode{
					Type:        "ExportDeclaration",
					Value:       exportName,
					StartPos:    Position{Line: lineNum, Column: 1},
					EndPos:      Position{Line: lineNum, Column: len(line)},
					Children:    []*ASTNode{},
					Attributes: map[string]interface{}{
						"name": exportName,
					},
				}
				nodes = append(nodes, node)
			}
		}
	}
	
	return nodes
}

// GetLanguage 返回解析器支持的语言
func (p *JavaScriptParser) GetLanguage() Language {
	return LanguageJavaScript
}

// GetVersion 返回解析器版本
func (p *JavaScriptParser) GetVersion() string {
	return "1.0.0"
}

// IsSupported 检查是否支持指定的文件扩展名
func (p *JavaScriptParser) IsSupported(filename string) bool {
	ext := strings.ToLower(filename)
	return strings.HasSuffix(ext, ".js") || 
		   strings.HasSuffix(ext, ".jsx") || 
		   strings.HasSuffix(ext, ".mjs") || 
		   strings.HasSuffix(ext, ".cjs")
}

// Validate 验证解析器配置
func (p *JavaScriptParser) Validate() error {
	// JavaScript解析器配置验证
	return nil
}

// SerializeAST 将AST序列化为JSON
func (p *JavaScriptParser) SerializeAST(ast *ASTNode) ([]byte, error) {
	return json.Marshal(ast)
}