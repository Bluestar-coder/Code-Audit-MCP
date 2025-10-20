package indexer

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// parseJavaScriptFile 解析 JavaScript 文件
func (s *IndexerService) parseJavaScriptFile(filePath string, astData []byte) (*FileIndex, error) {
	index := &FileIndex{
		FilePath:     filePath,
		Language:     "javascript",
		LastModified: time.Now(),
		Functions:    make(map[string]*Function),
		Classes:      make(map[string]*Class),
		Variables:    make(map[string]*Variable),
		CallGraph:    make(map[string][]string),
	}

	content := string(astData)
	
	// 提取导入语句
	index.Imports = s.extractJSImports(content)
	
	// 提取函数
	functions := s.extractJSFunctions(content, filePath)
	for _, function := range functions {
		index.Functions[function.ID] = function
	}
	
	// 提取类
	classes := s.extractJSClasses(content, filePath)
	for _, class := range classes {
		index.Classes[class.ID] = class
	}
	
	// 提取变量
	variables := s.extractJSVariables(content, filePath)
	for _, variable := range variables {
		index.Variables[variable.ID] = variable
	}

	return index, nil
}

// parseTypeScriptFile 解析 TypeScript 文件
func (s *IndexerService) parseTypeScriptFile(filePath string, astData []byte) (*FileIndex, error) {
	index := &FileIndex{
		FilePath:     filePath,
		Language:     "typescript",
		LastModified: time.Now(),
		Functions:    make(map[string]*Function),
		Classes:      make(map[string]*Class),
		Variables:    make(map[string]*Variable),
		CallGraph:    make(map[string][]string),
	}

	content := string(astData)
	
	// 提取导入语句
	index.Imports = s.extractTSImports(content)
	
	// 提取函数
	functions := s.extractTSFunctions(content, filePath)
	for _, function := range functions {
		index.Functions[function.ID] = function
	}
	
	// 提取类
	classes := s.extractTSClasses(content, filePath)
	for _, class := range classes {
		index.Classes[class.ID] = class
	}
	
	// 提取接口
	interfaces := s.extractTSInterfaces(content, filePath)
	for _, iface := range interfaces {
		index.Classes[iface.ID] = iface
	}
	
	// 提取变量
	variables := s.extractTSVariables(content, filePath)
	for _, variable := range variables {
		index.Variables[variable.ID] = variable
	}

	return index, nil
}

// parsePythonFile 解析 Python 文件
func (s *IndexerService) parsePythonFile(filePath string, astData []byte) (*FileIndex, error) {
	index := &FileIndex{
		FilePath:     filePath,
		Language:     "python",
		LastModified: time.Now(),
		Functions:    make(map[string]*Function),
		Classes:      make(map[string]*Class),
		Variables:    make(map[string]*Variable),
		CallGraph:    make(map[string][]string),
	}

	content := string(astData)
	
	// 提取导入语句
	index.Imports = s.extractPythonImports(content)
	
	// 提取函数
	functions := s.extractPythonFunctions(content, filePath)
	for _, function := range functions {
		index.Functions[function.ID] = function
	}
	
	// 提取类
	classes := s.extractPythonClasses(content, filePath)
	for _, class := range classes {
		index.Classes[class.ID] = class
	}
	
	// 提取变量
	variables := s.extractPythonVariables(content, filePath)
	for _, variable := range variables {
		index.Variables[variable.ID] = variable
	}

	return index, nil
}

// JavaScript 解析器

func (s *IndexerService) extractJSImports(content string) []string {
	var imports []string
	
	// import ... from '...'
	importRegex := regexp.MustCompile(`import\s+.*?\s+from\s+['"]([^'"]+)['"]`)
	matches := importRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			imports = append(imports, match[1])
		}
	}
	
	// require('...')
	requireRegex := regexp.MustCompile(`require\s*\(\s*['"]([^'"]+)['"]\s*\)`)
	matches = requireRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			imports = append(imports, match[1])
		}
	}
	
	return imports
}

func (s *IndexerService) extractJSFunctions(content, filePath string) []*Function {
	var functions []*Function
	
	// 函数声明: function name() {}
	funcRegex := regexp.MustCompile(`function\s+(\w+)\s*\(([^)]*)\)\s*\{`)
	matches := funcRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			name := match[1]
			params := match[2]
			line := s.getLineNumber(content, match[0])
			
			function := &Function{
				ID:        s.generateFunctionID(filePath, name, line),
				Name:      name,
				FilePath:  filePath,
				StartLine: line,
				Signature: fmt.Sprintf("function %s(%s)", name, params),
				Variables: make(map[string]string),
			}
			
			// 解析参数
			if params != "" {
				paramList := strings.Split(params, ",")
				for _, param := range paramList {
					param = strings.TrimSpace(param)
					if param != "" {
						function.Parameters = append(function.Parameters, Parameter{
							Name: param,
							Type: "any",
						})
					}
				}
			}
			
			functions = append(functions, function)
		}
	}
	
	// 箭头函数: const name = () => {}
	arrowRegex := regexp.MustCompile(`(?:const|let|var)\s+(\w+)\s*=\s*\(([^)]*)\)\s*=>\s*\{`)
	matches = arrowRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			name := match[1]
			params := match[2]
			line := s.getLineNumber(content, match[0])
			
			function := &Function{
				ID:        s.generateFunctionID(filePath, name, line),
				Name:      name,
				FilePath:  filePath,
				StartLine: line,
				Signature: fmt.Sprintf("const %s = (%s) => {}", name, params),
				Variables: make(map[string]string),
			}
			
			// 解析参数
			if params != "" {
				paramList := strings.Split(params, ",")
				for _, param := range paramList {
					param = strings.TrimSpace(param)
					if param != "" {
						function.Parameters = append(function.Parameters, Parameter{
							Name: param,
							Type: "any",
						})
					}
				}
			}
			
			functions = append(functions, function)
		}
	}
	
	return functions
}

func (s *IndexerService) extractJSClasses(content, filePath string) []*Class {
	var classes []*Class
	
	// class Name {}
	classRegex := regexp.MustCompile(`class\s+(\w+)(?:\s+extends\s+(\w+))?\s*\{`)
	matches := classRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			name := match[1]
			baseClass := ""
			if len(match) > 2 {
				baseClass = match[2]
			}
			line := s.getLineNumber(content, match[0])
			
			class := &Class{
				ID:        s.generateClassID(filePath, name, line),
				Name:      name,
				FilePath:  filePath,
				StartLine: line,
				BaseClass: baseClass,
				Methods:   make(map[string]*Method),
				Fields:    make(map[string]*Field),
			}
			
			classes = append(classes, class)
		}
	}
	
	return classes
}

func (s *IndexerService) extractJSVariables(content, filePath string) []*Variable {
	var variables []*Variable
	
	// var/let/const declarations
	varRegex := regexp.MustCompile(`(?:var|let|const)\s+(\w+)`)
	matches := varRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			name := match[1]
			line := s.getLineNumber(content, match[0])
			
			variable := &Variable{
				ID:       s.generateVariableID(filePath, name, line),
				Name:     name,
				Type:     "any",
				FilePath: filePath,
				Line:     line,
				Scope:    "global",
			}
			
			variables = append(variables, variable)
		}
	}
	
	return variables
}

// TypeScript 解析器

func (s *IndexerService) extractTSImports(content string) []string {
	var imports []string
	
	// import ... from '...'
	importRegex := regexp.MustCompile(`import\s+.*?\s+from\s+['"]([^'"]+)['"]`)
	matches := importRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			imports = append(imports, match[1])
		}
	}
	
	return imports
}

func (s *IndexerService) extractTSFunctions(content, filePath string) []*Function {
	var functions []*Function
	
	// function name(params: type): returnType {}
	funcRegex := regexp.MustCompile(`function\s+(\w+)\s*\(([^)]*)\)(?:\s*:\s*([^{]+))?\s*\{`)
	matches := funcRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			name := match[1]
			params := match[2]
			returnType := "void"
			if len(match) > 3 && match[3] != "" {
				returnType = strings.TrimSpace(match[3])
			}
			line := s.getLineNumber(content, match[0])
			
			function := &Function{
				ID:         s.generateFunctionID(filePath, name, line),
				Name:       name,
				FilePath:   filePath,
				StartLine:  line,
				ReturnType: returnType,
				Signature:  fmt.Sprintf("function %s(%s): %s", name, params, returnType),
				Variables:  make(map[string]string),
			}
			
			// 解析参数
			if params != "" {
				paramList := strings.Split(params, ",")
				for _, param := range paramList {
					param = strings.TrimSpace(param)
					if param != "" {
						parts := strings.Split(param, ":")
						paramName := strings.TrimSpace(parts[0])
						paramType := "any"
						if len(parts) > 1 {
							paramType = strings.TrimSpace(parts[1])
						}
						function.Parameters = append(function.Parameters, Parameter{
							Name: paramName,
							Type: paramType,
						})
					}
				}
			}
			
			functions = append(functions, function)
		}
	}
	
	return functions
}

func (s *IndexerService) extractTSClasses(content, filePath string) []*Class {
	var classes []*Class
	
	// class Name extends BaseClass implements Interface {}
	classRegex := regexp.MustCompile(`class\s+(\w+)(?:\s+extends\s+(\w+))?(?:\s+implements\s+([^{]+))?\s*\{`)
	matches := classRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			name := match[1]
			baseClass := ""
			if len(match) > 2 {
				baseClass = match[2]
			}
			interfaces := []string{}
			if len(match) > 3 && match[3] != "" {
				interfaceList := strings.Split(match[3], ",")
				for _, iface := range interfaceList {
					interfaces = append(interfaces, strings.TrimSpace(iface))
				}
			}
			line := s.getLineNumber(content, match[0])
			
			class := &Class{
				ID:         s.generateClassID(filePath, name, line),
				Name:       name,
				FilePath:   filePath,
				StartLine:  line,
				BaseClass:  baseClass,
				Interfaces: interfaces,
				Methods:    make(map[string]*Method),
				Fields:     make(map[string]*Field),
			}
			
			classes = append(classes, class)
		}
	}
	
	return classes
}

func (s *IndexerService) extractTSInterfaces(content, filePath string) []*Class {
	var interfaces []*Class
	
	// interface Name extends BaseInterface {}
	interfaceRegex := regexp.MustCompile(`interface\s+(\w+)(?:\s+extends\s+([^{]+))?\s*\{`)
	matches := interfaceRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			name := match[1]
			baseInterfaces := []string{}
			if len(match) > 2 && match[2] != "" {
				baseList := strings.Split(match[2], ",")
				for _, base := range baseList {
					baseInterfaces = append(baseInterfaces, strings.TrimSpace(base))
				}
			}
			line := s.getLineNumber(content, match[0])
			
			iface := &Class{
				ID:         s.generateClassID(filePath, name, line),
				Name:       name,
				FilePath:   filePath,
				StartLine:  line,
				Interfaces: baseInterfaces,
				Methods:    make(map[string]*Method),
				Fields:     make(map[string]*Field),
			}
			
			interfaces = append(interfaces, iface)
		}
	}
	
	return interfaces
}

func (s *IndexerService) extractTSVariables(content, filePath string) []*Variable {
	var variables []*Variable
	
	// var/let/const name: type = value
	varRegex := regexp.MustCompile(`(?:var|let|const)\s+(\w+)(?:\s*:\s*([^=\n]+))?`)
	matches := varRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			name := match[1]
			varType := "any"
			if len(match) > 2 && match[2] != "" {
				varType = strings.TrimSpace(match[2])
			}
			line := s.getLineNumber(content, match[0])
			
			variable := &Variable{
				ID:       s.generateVariableID(filePath, name, line),
				Name:     name,
				Type:     varType,
				FilePath: filePath,
				Line:     line,
				Scope:    "global",
			}
			
			variables = append(variables, variable)
		}
	}
	
	return variables
}

// Python 解析器

func (s *IndexerService) extractPythonImports(content string) []string {
	var imports []string
	
	// import module
	importRegex := regexp.MustCompile(`import\s+([^\n]+)`)
	matches := importRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			imports = append(imports, strings.TrimSpace(match[1]))
		}
	}
	
	// from module import ...
	fromRegex := regexp.MustCompile(`from\s+([^\s]+)\s+import`)
	matches = fromRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			imports = append(imports, match[1])
		}
	}
	
	return imports
}

func (s *IndexerService) extractPythonFunctions(content, filePath string) []*Function {
	var functions []*Function
	
	// def function_name(params):
	funcRegex := regexp.MustCompile(`def\s+(\w+)\s*\(([^)]*)\)(?:\s*->\s*([^:]+))?\s*:`)
	matches := funcRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			name := match[1]
			params := match[2]
			returnType := "None"
			if len(match) > 3 && match[3] != "" {
				returnType = strings.TrimSpace(match[3])
			}
			line := s.getLineNumber(content, match[0])
			
			function := &Function{
				ID:         s.generateFunctionID(filePath, name, line),
				Name:       name,
				FilePath:   filePath,
				StartLine:  line,
				ReturnType: returnType,
				Signature:  fmt.Sprintf("def %s(%s) -> %s:", name, params, returnType),
				Variables:  make(map[string]string),
			}
			
			// 解析参数
			if params != "" {
				paramList := strings.Split(params, ",")
				for _, param := range paramList {
					param = strings.TrimSpace(param)
					if param != "" && param != "self" {
						parts := strings.Split(param, ":")
						paramName := strings.TrimSpace(parts[0])
						paramType := "Any"
						if len(parts) > 1 {
							paramType = strings.TrimSpace(parts[1])
						}
						function.Parameters = append(function.Parameters, Parameter{
							Name: paramName,
							Type: paramType,
						})
					}
				}
			}
			
			functions = append(functions, function)
		}
	}
	
	return functions
}

func (s *IndexerService) extractPythonClasses(content, filePath string) []*Class {
	var classes []*Class
	
	// class Name(BaseClass):
	classRegex := regexp.MustCompile(`class\s+(\w+)(?:\s*\(([^)]*)\))?\s*:`)
	matches := classRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			name := match[1]
			baseClasses := []string{}
			if len(match) > 2 && match[2] != "" {
				baseList := strings.Split(match[2], ",")
				for _, base := range baseList {
					baseClasses = append(baseClasses, strings.TrimSpace(base))
				}
			}
			line := s.getLineNumber(content, match[0])
			
			baseClass := ""
			if len(baseClasses) > 0 {
				baseClass = baseClasses[0]
			}
			
			class := &Class{
				ID:         s.generateClassID(filePath, name, line),
				Name:       name,
				FilePath:   filePath,
				StartLine:  line,
				BaseClass:  baseClass,
				Interfaces: baseClasses[1:], // 其余作为接口
				Methods:    make(map[string]*Method),
				Fields:     make(map[string]*Field),
			}
			
			classes = append(classes, class)
		}
	}
	
	return classes
}

func (s *IndexerService) extractPythonVariables(content, filePath string) []*Variable {
	var variables []*Variable
	
	// variable_name = value 或 variable_name: type = value
	varRegex := regexp.MustCompile(`^(\w+)(?:\s*:\s*([^=\n]+))?\s*=`)
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		matches := varRegex.FindStringSubmatch(strings.TrimSpace(line))
		if len(matches) > 1 {
			name := matches[1]
			varType := "Any"
			if len(matches) > 2 && matches[2] != "" {
				varType = strings.TrimSpace(matches[2])
			}
			
			variable := &Variable{
				ID:       s.generateVariableID(filePath, name, i+1),
				Name:     name,
				Type:     varType,
				FilePath: filePath,
				Line:     i + 1,
				Scope:    "global",
			}
			
			variables = append(variables, variable)
		}
	}
	
	return variables
}

// 辅助方法

// getLineNumber 获取匹配文本在内容中的行号
func (s *IndexerService) getLineNumber(content, match string) int {
	index := strings.Index(content, match)
	if index == -1 {
		return 1
	}
	
	lines := strings.Count(content[:index], "\n")
	return lines + 1
}