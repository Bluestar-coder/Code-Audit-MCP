package indexer

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
)

// IndexerService 代码索引服务
type IndexerService struct {
	db       *badger.DB
	mu       sync.RWMutex
	fileSet  *token.FileSet
	indexes  map[string]*FileIndex // 文件路径 -> 索引
}

// FileIndex 文件索引
type FileIndex struct {
	FilePath    string                 `json:"file_path"`
	Language    string                 `json:"language"`
	LastModified time.Time             `json:"last_modified"`
	Functions   map[string]*Function   `json:"functions"`
	Classes     map[string]*Class      `json:"classes"`
	Variables   map[string]*Variable   `json:"variables"`
	Imports     []string               `json:"imports"`
	CallGraph   map[string][]string    `json:"call_graph"` // 函数ID -> 调用的函数ID列表
}

// Function 函数信息
type Function struct {
	ID                   string            `json:"id"`
	Name                 string            `json:"name"`
	FilePath             string            `json:"file_path"`
	StartLine            int               `json:"start_line"`
	EndLine              int               `json:"end_line"`
	Signature            string            `json:"signature"`
	CyclomaticComplexity int               `json:"cyclomatic_complexity"`
	Parameters           []Parameter       `json:"parameters"`
	ReturnType           string            `json:"return_type"`
	Calls                []string          `json:"calls"`        // 调用的函数
	CalledBy             []string          `json:"called_by"`    // 被谁调用
	Variables            map[string]string `json:"variables"`    // 局部变量
	Comments             []string          `json:"comments"`
}

// Parameter 参数信息
type Parameter struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// Class 类信息
type Class struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	FilePath   string            `json:"file_path"`
	StartLine  int               `json:"start_line"`
	EndLine    int               `json:"end_line"`
	BaseClass  string            `json:"base_class"`
	Interfaces []string          `json:"interfaces"`
	Methods    map[string]*Method `json:"methods"`
	Fields     map[string]*Field  `json:"fields"`
	Comments   []string          `json:"comments"`
}

// Method 方法信息
type Method struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	IsStatic   bool        `json:"is_static"`
	IsPrivate  bool        `json:"is_private"`
	IsPublic   bool        `json:"is_public"`
	ReturnType string      `json:"return_type"`
	Parameters []Parameter `json:"parameters"`
	StartLine  int         `json:"start_line"`
	EndLine    int         `json:"end_line"`
}

// Field 字段信息
type Field struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	IsStatic  bool   `json:"is_static"`
	IsPrivate bool   `json:"is_private"`
	IsPublic  bool   `json:"is_public"`
	Line      int    `json:"line"`
}

// Variable 变量信息
type Variable struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	FilePath string `json:"file_path"`
	Line     int    `json:"line"`
	Scope    string `json:"scope"` // global, function, class
}

// NewIndexerService 创建新的索引服务
func NewIndexerService(dbPath string) (*IndexerService, error) {
	opts := badger.DefaultOptions(dbPath)
	opts.Logger = nil // 禁用日志
	
	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	return &IndexerService{
		db:      db,
		fileSet: token.NewFileSet(),
		indexes: make(map[string]*FileIndex),
	}, nil
}

// Close 关闭服务
func (s *IndexerService) Close() error {
	return s.db.Close()
}

// BuildIndex 构建文件索引
func (s *IndexerService) BuildIndex(filePath, language string, astData []byte, incremental bool) (*IndexResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("🔨 Building index for %s (%s)", filePath, language)

	// 检查是否需要增量更新
	if incremental {
		if existing, exists := s.indexes[filePath]; exists {
			// 检查文件是否有变化
			if !s.hasFileChanged(filePath, existing.LastModified) {
				log.Printf("📋 File %s unchanged, skipping index", filePath)
				return &IndexResult{
					Success:          true,
					IndexID:          s.generateIndexID(filePath),
					FunctionsIndexed: len(existing.Functions),
					ClassesIndexed:   len(existing.Classes),
					VariablesIndexed: len(existing.Variables),
				}, nil
			}
		}
	}

	// 根据语言选择解析器
	var index *FileIndex
	var err error

	switch strings.ToLower(language) {
	case "go", "golang":
		index, err = s.parseGoFile(filePath, astData)
	case "javascript", "js":
		index, err = s.parseJavaScriptFile(filePath, astData)
	case "typescript", "ts":
		index, err = s.parseTypeScriptFile(filePath, astData)
	case "python", "py":
		index, err = s.parsePythonFile(filePath, astData)
	default:
		return nil, fmt.Errorf("unsupported language: %s", language)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %v", err)
	}

	// 保存索引到内存
	s.indexes[filePath] = index

	// 保存索引到数据库
	if err := s.saveIndexToDB(filePath, index); err != nil {
		log.Printf("⚠️ Failed to save index to DB: %v", err)
	}

	// 更新调用关系
	s.updateCallGraph(index)

	result := &IndexResult{
		Success:          true,
		IndexID:          s.generateIndexID(filePath),
		FunctionsIndexed: len(index.Functions),
		ClassesIndexed:   len(index.Classes),
		VariablesIndexed: len(index.Variables),
	}

	log.Printf("✅ Index built: %d functions, %d classes, %d variables", 
		result.FunctionsIndexed, result.ClassesIndexed, result.VariablesIndexed)

	return result, nil
}

// IndexResult 索引结果
type IndexResult struct {
	Success          bool   `json:"success"`
	IndexID          string `json:"index_id"`
	FunctionsIndexed int    `json:"functions_indexed"`
	ClassesIndexed   int    `json:"classes_indexed"`
	VariablesIndexed int    `json:"variables_indexed"`
	ErrorMessage     string `json:"error_message,omitempty"`
}

// parseGoFile 解析 Go 文件
func (s *IndexerService) parseGoFile(filePath string, astData []byte) (*FileIndex, error) {
	// 解析 AST
	file, err := parser.ParseFile(s.fileSet, filePath, astData, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Go file: %v", err)
	}

	index := &FileIndex{
		FilePath:     filePath,
		Language:     "go",
		LastModified: time.Now(),
		Functions:    make(map[string]*Function),
		Classes:      make(map[string]*Class),
		Variables:    make(map[string]*Variable),
		CallGraph:    make(map[string][]string),
	}

	// 提取导入
	for _, imp := range file.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		index.Imports = append(index.Imports, path)
	}

	// 遍历 AST 节点
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			if node.Name != nil {
				function := s.extractGoFunction(node, filePath)
				index.Functions[function.ID] = function
			}
		case *ast.GenDecl:
			// 处理类型声明（结构体、接口等）
			for _, spec := range node.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if structType, ok := s.Type.(*ast.StructType); ok {
						class := s.extractGoStruct(s, structType, filePath)
						index.Classes[class.ID] = class
					}
				case *ast.ValueSpec:
					// 处理变量声明
					for _, name := range s.Names {
						variable := s.extractGoVariable(name, s, filePath)
						index.Variables[variable.ID] = variable
					}
				}
			}
		}
		return true
	})

	return index, nil
}

// extractGoFunction 提取 Go 函数信息
func (s *IndexerService) extractGoFunction(node *ast.FuncDecl, filePath string) *Function {
	pos := s.fileSet.Position(node.Pos())
	end := s.fileSet.Position(node.End())
	
	function := &Function{
		ID:        s.generateFunctionID(filePath, node.Name.Name, pos.Line),
		Name:      node.Name.Name,
		FilePath:  filePath,
		StartLine: pos.Line,
		EndLine:   end.Line,
		Variables: make(map[string]string),
	}

	// 提取参数
	if node.Type.Params != nil {
		for _, field := range node.Type.Params.List {
			paramType := s.extractTypeString(field.Type)
			for _, name := range field.Names {
				function.Parameters = append(function.Parameters, Parameter{
					Name: name.Name,
					Type: paramType,
				})
			}
		}
	}

	// 提取返回类型
	if node.Type.Results != nil {
		var returnTypes []string
		for _, field := range node.Type.Results.List {
			returnTypes = append(returnTypes, s.extractTypeString(field.Type))
		}
		function.ReturnType = strings.Join(returnTypes, ", ")
	}

	// 生成函数签名
	function.Signature = s.generateGoFunctionSignature(node)

	// 计算圈复杂度
	function.CyclomaticComplexity = s.calculateCyclomaticComplexity(node)

	// 提取函数调用
	function.Calls = s.extractFunctionCalls(node)

	return function
}

// extractGoStruct 提取 Go 结构体信息
func (s *IndexerService) extractGoStruct(typeSpec *ast.TypeSpec, structType *ast.StructType, filePath string) *Class {
	pos := s.fileSet.Position(typeSpec.Pos())
	end := s.fileSet.Position(typeSpec.End())

	class := &Class{
		ID:        s.generateClassID(filePath, typeSpec.Name.Name, pos.Line),
		Name:      typeSpec.Name.Name,
		FilePath:  filePath,
		StartLine: pos.Line,
		EndLine:   end.Line,
		Methods:   make(map[string]*Method),
		Fields:    make(map[string]*Field),
	}

	// 提取字段
	for _, field := range structType.Fields.List {
		fieldType := s.extractTypeString(field.Type)
		for _, name := range field.Names {
			fieldPos := s.fileSet.Position(name.Pos())
			class.Fields[name.Name] = &Field{
				Name:     name.Name,
				Type:     fieldType,
				IsPublic: ast.IsExported(name.Name),
				Line:     fieldPos.Line,
			}
		}
	}

	return class
}

// extractGoVariable 提取 Go 变量信息
func (s *IndexerService) extractGoVariable(name *ast.Ident, spec *ast.ValueSpec, filePath string) *Variable {
	pos := s.fileSet.Position(name.Pos())
	
	variable := &Variable{
		ID:       s.generateVariableID(filePath, name.Name, pos.Line),
		Name:     name.Name,
		FilePath: filePath,
		Line:     pos.Line,
		Scope:    "global",
	}

	// 提取类型
	if spec.Type != nil {
		variable.Type = s.extractTypeString(spec.Type)
	}

	return variable
}

// 辅助方法
func (s *IndexerService) extractTypeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return s.extractTypeString(t.X) + "." + t.Sel.Name
	case *ast.StarExpr:
		return "*" + s.extractTypeString(t.X)
	case *ast.ArrayType:
		return "[]" + s.extractTypeString(t.Elt)
	case *ast.MapType:
		return "map[" + s.extractTypeString(t.Key) + "]" + s.extractTypeString(t.Value)
	default:
		return "unknown"
	}
}

func (s *IndexerService) generateGoFunctionSignature(node *ast.FuncDecl) string {
	var parts []string
	parts = append(parts, "func")
	
	if node.Recv != nil {
		// 方法
		recv := s.extractTypeString(node.Recv.List[0].Type)
		parts = append(parts, "("+recv+")")
	}
	
	parts = append(parts, node.Name.Name+"(")
	
	// 参数
	var params []string
	if node.Type.Params != nil {
		for _, field := range node.Type.Params.List {
			paramType := s.extractTypeString(field.Type)
			for _, name := range field.Names {
				params = append(params, name.Name+" "+paramType)
			}
		}
	}
	parts = append(parts, strings.Join(params, ", ")+")")
	
	// 返回类型
	if node.Type.Results != nil {
		var returns []string
		for _, field := range node.Type.Results.List {
			returns = append(returns, s.extractTypeString(field.Type))
		}
		if len(returns) == 1 {
			parts = append(parts, " "+returns[0])
		} else {
			parts = append(parts, " ("+strings.Join(returns, ", ")+")")
		}
	}
	
	return strings.Join(parts, "")
}

func (s *IndexerService) calculateCyclomaticComplexity(node *ast.FuncDecl) int {
	complexity := 1 // 基础复杂度
	
	ast.Inspect(node, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.SwitchStmt, *ast.TypeSwitchStmt:
			complexity++
		case *ast.CaseClause:
			complexity++
		}
		return true
	})
	
	return complexity
}

func (s *IndexerService) extractFunctionCalls(node *ast.FuncDecl) []string {
	var calls []string
	
	ast.Inspect(node, func(n ast.Node) bool {
		if callExpr, ok := n.(*ast.CallExpr); ok {
			if ident, ok := callExpr.Fun.(*ast.Ident); ok {
				calls = append(calls, ident.Name)
			} else if sel, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
				calls = append(calls, sel.Sel.Name)
			}
		}
		return true
	})
	
	return calls
}

// 生成ID的辅助方法
func (s *IndexerService) generateIndexID(filePath string) string {
	hash := md5.Sum([]byte(filePath + time.Now().String()))
	return fmt.Sprintf("idx_%x", hash[:8])
}

func (s *IndexerService) generateFunctionID(filePath, name string, line int) string {
	hash := md5.Sum([]byte(fmt.Sprintf("%s:%s:%d", filePath, name, line)))
	return fmt.Sprintf("func_%x", hash[:8])
}

func (s *IndexerService) generateClassID(filePath, name string, line int) string {
	hash := md5.Sum([]byte(fmt.Sprintf("%s:%s:%d", filePath, name, line)))
	return fmt.Sprintf("class_%x", hash[:8])
}

func (s *IndexerService) generateVariableID(filePath, name string, line int) string {
	hash := md5.Sum([]byte(fmt.Sprintf("%s:%s:%d", filePath, name, line)))
	return fmt.Sprintf("var_%x", hash[:8])
}

// hasFileChanged 检查文件是否有变化
func (s *IndexerService) hasFileChanged(filePath string, lastModified time.Time) bool {
	// 这里可以实现文件修改时间检查
	// 简化实现，总是返回 true
	return true
}

// saveIndexToDB 保存索引到数据库
func (s *IndexerService) saveIndexToDB(filePath string, index *FileIndex) error {
	return s.db.Update(func(txn *badger.Txn) error {
		data, err := json.Marshal(index)
		if err != nil {
			return err
		}
		
		key := fmt.Sprintf("index:%s", filePath)
		return txn.Set([]byte(key), data)
	})
}

// updateCallGraph 更新调用关系图
func (s *IndexerService) updateCallGraph(index *FileIndex) {
	// 构建调用关系
	for funcID, function := range index.Functions {
		for _, call := range function.Calls {
			// 查找被调用的函数
			for targetID, target := range index.Functions {
				if target.Name == call {
					index.CallGraph[funcID] = append(index.CallGraph[funcID], targetID)
					target.CalledBy = append(target.CalledBy, funcID)
				}
			}
		}
	}
}