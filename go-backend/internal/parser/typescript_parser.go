package parser

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"time"
)

// TypeScriptParser 实现TypeScript语言的AST解析器
type TypeScriptParser struct{}

// NewTypeScriptParser 创建新的TypeScript解析器实例
func NewTypeScriptParser() *TypeScriptParser {
	return &TypeScriptParser{}
}

// Parse 解析TypeScript源代码
func (p *TypeScriptParser) Parse(ctx context.Context, content []byte, filePath string, options ParseOptions) (*ParseResult, error) {
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
	var functions, classes, variables, interfaces, types int
	
	for i, line := range lines {
		lineNum := i + 1
		trimmedLine := strings.TrimSpace(line)
		
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "//") {
			continue
		}
		
		// 解析TypeScript接口声明
		if nodes := p.parseInterfaceDeclaration(line, lineNum); len(nodes) > 0 {
			root.Children = append(root.Children, nodes...)
			interfaces += len(nodes)
		}
		
		// 解析TypeScript类型别名
		if nodes := p.parseTypeAlias(line, lineNum); len(nodes) > 0 {
			root.Children = append(root.Children, nodes...)
			types += len(nodes)
		}
		
		// 解析函数声明（包含类型注解）
		if nodes := p.parseFunctionDeclaration(line, lineNum); len(nodes) > 0 {
			root.Children = append(root.Children, nodes...)
			functions += len(nodes)
		}
		
		// 解析类声明（包含泛型和修饰符）
		if nodes := p.parseClassDeclaration(line, lineNum); len(nodes) > 0 {
			root.Children = append(root.Children, nodes...)
			classes += len(nodes)
		}
		
		// 解析变量声明（包含类型注解）
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
		
		// 解析枚举声明
		if nodes := p.parseEnumDeclaration(line, lineNum); len(nodes) > 0 {
			root.Children = append(root.Children, nodes...)
		}
		
		// 解析命名空间声明
		if nodes := p.parseNamespaceDeclaration(line, lineNum); len(nodes) > 0 {
			root.Children = append(root.Children, nodes...)
		}
	}
	
	parseTime := time.Since(startTime)
	
	metadata := ParseMetadata{
		ParseTime:       parseTime,
		TotalLines:      len(lines),
		TotalNodes:      p.countNodes(root),
		TotalFunctions:  functions,
		TotalClasses:    classes,
		TotalVariables:  variables,
		LanguageVersion: "TypeScript 5.0",
		ParserVersion:   "1.0.0",
	}
	
	return &ParseResult{
		AST:      root,
		Errors:   errors,
		Metadata: metadata,
		Language: LanguageTypeScript,
		FilePath: filePath,
		Success:  len(errors) == 0,
	}, nil
}

// parseInterfaceDeclaration 解析TypeScript接口声明
func (p *TypeScriptParser) parseInterfaceDeclaration(line string, lineNum int) []*ASTNode {
	// 匹配接口声明: interface Name<T> extends Base
	interfaceRegex := regexp.MustCompile(`^\s*(?:export\s+)?interface\s+([A-Za-z_$][A-Za-z0-9_$]*)\s*(?:<[^>]*>)?\s*(?:extends\s+[^{]+)?\s*\{?`)
	
	if matches := interfaceRegex.FindStringSubmatch(line); len(matches) > 1 {
		interfaceName := matches[1]
		
		node := &ASTNode{
			Type:     "InterfaceDeclaration",
			Value:    interfaceName,
			StartPos: Position{Line: lineNum, Column: 1},
			EndPos:   Position{Line: lineNum, Column: len(line)},
			Children: []*ASTNode{},
			Attributes: map[string]interface{}{
				"name":       interfaceName,
				"exported":   strings.Contains(line, "export"),
				"generic":    strings.Contains(line, "<"),
				"extends":    strings.Contains(line, "extends"),
			},
		}
		
		return []*ASTNode{node}
	}
	
	return nil
}

// parseTypeAlias 解析TypeScript类型别名
func (p *TypeScriptParser) parseTypeAlias(line string, lineNum int) []*ASTNode {
	// 匹配类型别名: type Name<T> = Type
	typeRegex := regexp.MustCompile(`^\s*(?:export\s+)?type\s+([A-Za-z_$][A-Za-z0-9_$]*)\s*(?:<[^>]*>)?\s*=`)
	
	if matches := typeRegex.FindStringSubmatch(line); len(matches) > 1 {
		typeName := matches[1]
		
		node := &ASTNode{
			Type:     "TypeAlias",
			Value:    typeName,
			StartPos: Position{Line: lineNum, Column: 1},
			EndPos:   Position{Line: lineNum, Column: len(line)},
			Children: []*ASTNode{},
			Attributes: map[string]interface{}{
				"name":     typeName,
				"exported": strings.Contains(line, "export"),
				"generic":  strings.Contains(line, "<"),
			},
		}
		
		return []*ASTNode{node}
	}
	
	return nil
}

// parseFunctionDeclaration 解析TypeScript函数声明（包含类型注解）
func (p *TypeScriptParser) parseFunctionDeclaration(line string, lineNum int) []*ASTNode {
	// 匹配函数声明: function name<T>(param: Type): ReturnType
	funcRegex := regexp.MustCompile(`^\s*(?:export\s+)?(?:async\s+)?function\s+([A-Za-z_$][A-Za-z0-9_$]*)\s*(?:<[^>]*>)?\s*\([^)]*\)\s*(?::\s*[^{]+)?\s*\{?`)
	
	// 匹配箭头函数: const name = (param: Type): ReturnType => 
	arrowRegex := regexp.MustCompile(`^\s*(?:export\s+)?(?:const|let|var)\s+([A-Za-z_$][A-Za-z0-9_$]*)\s*=\s*(?:async\s+)?\([^)]*\)\s*(?::\s*[^=]+)?\s*=>`)
	
	// 匹配方法声明: methodName<T>(param: Type): ReturnType
	methodRegex := regexp.MustCompile(`^\s*(?:public|private|protected|static)?\s*(?:async\s+)?([A-Za-z_$][A-Za-z0-9_$]*)\s*(?:<[^>]*>)?\s*\([^)]*\)\s*(?::\s*[^{]+)?\s*\{?`)
	
	var funcName string
	var funcType string
	
	if matches := funcRegex.FindStringSubmatch(line); len(matches) > 1 {
		funcName = matches[1]
		funcType = "FunctionDeclaration"
	} else if matches := arrowRegex.FindStringSubmatch(line); len(matches) > 1 {
		funcName = matches[1]
		funcType = "ArrowFunction"
	} else if matches := methodRegex.FindStringSubmatch(line); len(matches) > 1 {
		funcName = matches[1]
		funcType = "MethodDefinition"
	}
	
	if funcName != "" {
		node := &ASTNode{
			Type:     funcType,
			Value:    funcName,
			StartPos: Position{Line: lineNum, Column: 1},
			EndPos:   Position{Line: lineNum, Column: len(line)},
			Children: []*ASTNode{},
			Attributes: map[string]interface{}{
				"name":       funcName,
				"async":      strings.Contains(line, "async"),
				"exported":   strings.Contains(line, "export"),
				"generic":    strings.Contains(line, "<"),
				"typed":      strings.Contains(line, ":"),
				"visibility": p.extractVisibility(line),
			},
		}
		
		return []*ASTNode{node}
	}
	
	return nil
}

// parseClassDeclaration 解析TypeScript类声明（包含泛型和修饰符）
func (p *TypeScriptParser) parseClassDeclaration(line string, lineNum int) []*ASTNode {
	// 匹配类声明: class Name<T> extends Base implements Interface
	classRegex := regexp.MustCompile(`^\s*(?:export\s+)?(?:abstract\s+)?class\s+([A-Za-z_$][A-Za-z0-9_$]*)\s*(?:<[^>]*>)?\s*(?:extends\s+[A-Za-z_$][A-Za-z0-9_$]*\s*(?:<[^>]*>)?)?\s*(?:implements\s+[^{]+)?\s*\{?`)
	
	if matches := classRegex.FindStringSubmatch(line); len(matches) > 1 {
		className := matches[1]
		
		node := &ASTNode{
			Type:     "ClassDeclaration",
			Value:    className,
			StartPos: Position{Line: lineNum, Column: 1},
			EndPos:   Position{Line: lineNum, Column: len(line)},
			Children: []*ASTNode{},
			Attributes: map[string]interface{}{
				"name":        className,
				"exported":    strings.Contains(line, "export"),
				"abstract":    strings.Contains(line, "abstract"),
				"generic":     strings.Contains(line, "<"),
				"extends":     strings.Contains(line, "extends"),
				"implements":  strings.Contains(line, "implements"),
			},
		}
		
		return []*ASTNode{node}
	}
	
	return nil
}

// parseVariableDeclaration 解析TypeScript变量声明（包含类型注解）
func (p *TypeScriptParser) parseVariableDeclaration(line string, lineNum int) []*ASTNode {
	// 匹配变量声明: const/let/var name: Type = value
	varRegex := regexp.MustCompile(`^\s*(?:export\s+)?(?:const|let|var)\s+([A-Za-z_$][A-Za-z0-9_$]*)\s*(?::\s*[^=]+)?\s*=`)
	
	if matches := varRegex.FindStringSubmatch(line); len(matches) > 1 {
		varName := matches[1]
		
		node := &ASTNode{
			Type:     "VariableDeclaration",
			Value:    varName,
			StartPos: Position{Line: lineNum, Column: 1},
			EndPos:   Position{Line: lineNum, Column: len(line)},
			Children: []*ASTNode{},
			Attributes: map[string]interface{}{
				"name":     varName,
				"exported": strings.Contains(line, "export"),
				"typed":    strings.Contains(line, ":"),
				"kind":     p.extractVariableKind(line),
			},
		}
		
		return []*ASTNode{node}
	}
	
	return nil
}

// parseImportStatement 解析TypeScript导入语句
func (p *TypeScriptParser) parseImportStatement(line string, lineNum int) []*ASTNode {
	// 匹配导入语句: import { name } from 'module' 或 import type { Type } from 'module'
	importRegex := regexp.MustCompile(`^\s*import\s+(?:type\s+)?(?:\{[^}]*\}|\*\s+as\s+\w+|\w+)\s+from\s+['"][^'"]+['"]`)
	
	if importRegex.MatchString(line) {
		node := &ASTNode{
			Type:     "ImportDeclaration",
			Value:    strings.TrimSpace(line),
			StartPos: Position{Line: lineNum, Column: 1},
			EndPos:   Position{Line: lineNum, Column: len(line)},
			Children: []*ASTNode{},
			Attributes: map[string]interface{}{
				"typeOnly": strings.Contains(line, "import type"),
				"namespace": strings.Contains(line, "* as"),
			},
		}
		
		return []*ASTNode{node}
	}
	
	return nil
}

// parseExportStatement 解析TypeScript导出语句
func (p *TypeScriptParser) parseExportStatement(line string, lineNum int) []*ASTNode {
	// 匹配导出语句: export { name } 或 export type { Type }
	exportRegex := regexp.MustCompile(`^\s*export\s+(?:type\s+)?\{[^}]*\}(?:\s+from\s+['"][^'"]+['"])?`)
	
	if exportRegex.MatchString(line) {
		node := &ASTNode{
			Type:     "ExportDeclaration",
			Value:    strings.TrimSpace(line),
			StartPos: Position{Line: lineNum, Column: 1},
			EndPos:   Position{Line: lineNum, Column: len(line)},
			Children: []*ASTNode{},
			Attributes: map[string]interface{}{
				"typeOnly": strings.Contains(line, "export type"),
				"reExport": strings.Contains(line, "from"),
			},
		}
		
		return []*ASTNode{node}
	}
	
	return nil
}

// parseEnumDeclaration 解析TypeScript枚举声明
func (p *TypeScriptParser) parseEnumDeclaration(line string, lineNum int) []*ASTNode {
	// 匹配枚举声明: enum Name { ... } 或 const enum Name { ... }
	enumRegex := regexp.MustCompile(`^\s*(?:export\s+)?(?:const\s+)?enum\s+([A-Za-z_$][A-Za-z0-9_$]*)\s*\{?`)
	
	if matches := enumRegex.FindStringSubmatch(line); len(matches) > 1 {
		enumName := matches[1]
		
		node := &ASTNode{
			Type:     "EnumDeclaration",
			Value:    enumName,
			StartPos: Position{Line: lineNum, Column: 1},
			EndPos:   Position{Line: lineNum, Column: len(line)},
			Children: []*ASTNode{},
			Attributes: map[string]interface{}{
				"name":     enumName,
				"exported": strings.Contains(line, "export"),
				"const":    strings.Contains(line, "const enum"),
			},
		}
		
		return []*ASTNode{node}
	}
	
	return nil
}

// parseNamespaceDeclaration 解析TypeScript命名空间声明
func (p *TypeScriptParser) parseNamespaceDeclaration(line string, lineNum int) []*ASTNode {
	// 匹配命名空间声明: namespace Name { ... } 或 module Name { ... }
	namespaceRegex := regexp.MustCompile(`^\s*(?:export\s+)?(?:namespace|module)\s+([A-Za-z_$][A-Za-z0-9_$.]*)\s*\{?`)
	
	if matches := namespaceRegex.FindStringSubmatch(line); len(matches) > 1 {
		namespaceName := matches[1]
		
		node := &ASTNode{
			Type:     "NamespaceDeclaration",
			Value:    namespaceName,
			StartPos: Position{Line: lineNum, Column: 1},
			EndPos:   Position{Line: lineNum, Column: len(line)},
			Children: []*ASTNode{},
			Attributes: map[string]interface{}{
				"name":     namespaceName,
				"exported": strings.Contains(line, "export"),
			},
		}
		
		return []*ASTNode{node}
	}
	
	return nil
}

// extractVisibility 提取可见性修饰符
func (p *TypeScriptParser) extractVisibility(line string) string {
	if strings.Contains(line, "private") {
		return "private"
	} else if strings.Contains(line, "protected") {
		return "protected"
	} else if strings.Contains(line, "public") {
		return "public"
	}
	return "public" // 默认为public
}

// extractVariableKind 提取变量声明类型
func (p *TypeScriptParser) extractVariableKind(line string) string {
	if strings.Contains(line, "const") {
		return "const"
	} else if strings.Contains(line, "let") {
		return "let"
	} else if strings.Contains(line, "var") {
		return "var"
	}
	return "unknown"
}

// countNodes 递归计算AST节点数量
func (p *TypeScriptParser) countNodes(node *ASTNode) int {
	count := 1
	for _, child := range node.Children {
		count += p.countNodes(child)
	}
	return count
}

// GetLanguage 返回解析器支持的语言
func (p *TypeScriptParser) GetLanguage() Language {
	return LanguageTypeScript
}

// GetVersion 返回解析器版本
func (p *TypeScriptParser) GetVersion() string {
	return "1.0.0"
}

// IsSupported 检查文件是否被支持
func (p *TypeScriptParser) IsSupported(filePath string) bool {
	return strings.HasSuffix(strings.ToLower(filePath), ".ts") ||
		   strings.HasSuffix(strings.ToLower(filePath), ".tsx")
}

// Validate 验证解析器配置
func (p *TypeScriptParser) Validate() error {
	return nil
}

// SerializeAST 序列化AST为JSON
func (p *TypeScriptParser) SerializeAST(ast *ASTNode) ([]byte, error) {
	return json.Marshal(ast)
}