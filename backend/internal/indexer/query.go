package indexer

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/dgraph-io/badger/v4"
)

// QueryFunctionRequest 查询函数请求
type QueryFunctionRequest struct {
	FunctionName string `json:"function_name"`
	FilePath     string `json:"file_path,omitempty"`
	ExactMatch   bool   `json:"exact_match"`
}

// QueryFunctionResponse 查询函数响应
type QueryFunctionResponse struct {
	Functions  []*Function `json:"functions"`
	TotalCount int         `json:"total_count"`
}

// QueryClassRequest 查询类请求
type QueryClassRequest struct {
	ClassName  string `json:"class_name"`
	FilePath   string `json:"file_path,omitempty"`
	ExactMatch bool   `json:"exact_match"`
}

// QueryClassResponse 查询类响应
type QueryClassResponse struct {
	Classes    []*Class `json:"classes"`
	TotalCount int      `json:"total_count"`
}

// QueryCallersRequest 查询调用者请求
type QueryCallersRequest struct {
	FunctionID string `json:"function_id"`
	MaxDepth   int    `json:"max_depth"`
}

// QueryCallersResponse 查询调用者响应
type QueryCallersResponse struct {
	Callers    []*CallInfo `json:"callers"`
	TotalCount int         `json:"total_count"`
}

// QueryCalleesRequest 查询被调用者请求
type QueryCalleesRequest struct {
	FunctionID string `json:"function_id"`
	MaxDepth   int    `json:"max_depth"`
}

// QueryCalleesResponse 查询被调用者响应
type QueryCalleesResponse struct {
	Callees    []*CallInfo `json:"callees"`
	TotalCount int         `json:"total_count"`
}

// CallInfo 调用信息
type CallInfo struct {
	CallerID   string `json:"caller_id"`
	CallerName string `json:"caller_name"`
	CalleeID   string `json:"callee_id"`
	CalleeName string `json:"callee_name"`
	CallLine   int    `json:"call_line"`
	CallType   string `json:"call_type"` // direct, indirect
	FilePath   string `json:"file_path"`
}

// SearchSymbolRequest 搜索符号请求
type SearchSymbolRequest struct {
	Pattern    string `json:"pattern"`
	SymbolType string `json:"symbol_type"` // function, class, variable, all
	MaxResults int    `json:"max_results"`
}

// SearchSymbolResponse 搜索符号响应
type SearchSymbolResponse struct {
	SymbolID       string  `json:"symbol_id"`
	SymbolName     string  `json:"symbol_name"`
	SymbolType     string  `json:"symbol_type"`
	FilePath       string  `json:"file_path"`
	LineNumber     int     `json:"line_number"`
	RelevanceScore float64 `json:"relevance_score"`
	Context        string  `json:"context,omitempty"`
}

// QueryFunction 查询函数
func (s *IndexerService) QueryFunction(req *QueryFunctionRequest) (*QueryFunctionResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var functions []*Function

	// 从内存索引中查询
	for _, index := range s.indexes {
		// 如果指定了文件路径，只在该文件中查询
		if req.FilePath != "" && index.FilePath != req.FilePath {
			continue
		}

		for _, function := range index.Functions {
			if s.matchFunction(function, req.FunctionName, req.ExactMatch) {
				functions = append(functions, function)
			}
		}
	}

	// 如果内存中没有找到，从数据库中查询
	if len(functions) == 0 {
		dbFunctions, err := s.queryFunctionFromDB(req)
		if err == nil {
			functions = dbFunctions
		}
	}

	// 按相关性排序
	sort.Slice(functions, func(i, j int) bool {
		return s.calculateFunctionRelevance(functions[i], req.FunctionName) >
			s.calculateFunctionRelevance(functions[j], req.FunctionName)
	})

	return &QueryFunctionResponse{
		Functions:  functions,
		TotalCount: len(functions),
	}, nil
}

// QueryClass 查询类
func (s *IndexerService) QueryClass(req *QueryClassRequest) (*QueryClassResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var classes []*Class

	// 从内存索引中查询
	for _, index := range s.indexes {
		// 如果指定了文件路径，只在该文件中查询
		if req.FilePath != "" && index.FilePath != req.FilePath {
			continue
		}

		for _, class := range index.Classes {
			if s.matchClass(class, req.ClassName, req.ExactMatch) {
				classes = append(classes, class)
			}
		}
	}

	// 如果内存中没有找到，从数据库中查询
	if len(classes) == 0 {
		dbClasses, err := s.queryClassFromDB(req)
		if err == nil {
			classes = dbClasses
		}
	}

	// 按相关性排序
	sort.Slice(classes, func(i, j int) bool {
		return s.calculateClassRelevance(classes[i], req.ClassName) >
			s.calculateClassRelevance(classes[j], req.ClassName)
	})

	return &QueryClassResponse{
		Classes:    classes,
		TotalCount: len(classes),
	}, nil
}

// QueryCallers 查询调用者
func (s *IndexerService) QueryCallers(req *QueryCallersRequest) (*QueryCallersResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var callers []*CallInfo
	visited := make(map[string]bool)

	// 递归查找调用者
	s.findCallersRecursive(req.FunctionID, req.MaxDepth, 0, visited, &callers)

	return &QueryCallersResponse{
		Callers:    callers,
		TotalCount: len(callers),
	}, nil
}

// QueryCallees 查询被调用者
func (s *IndexerService) QueryCallees(req *QueryCalleesRequest) (*QueryCalleesResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var callees []*CallInfo
	visited := make(map[string]bool)

	// 递归查找被调用者
	s.findCalleesRecursive(req.FunctionID, req.MaxDepth, 0, visited, &callees)

	return &QueryCalleesResponse{
		Callees:    callees,
		TotalCount: len(callees),
	}, nil
}

// SearchSymbol 搜索符号
func (s *IndexerService) SearchSymbol(req *SearchSymbolRequest) ([]*SearchSymbolResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*SearchSymbolResponse
	pattern := req.Pattern

	// 编译正则表达式
	var regex *regexp.Regexp
	var err error
	if strings.HasPrefix(pattern, "/") && strings.HasSuffix(pattern, "/") {
		// 正则表达式模式
		regexPattern := pattern[1 : len(pattern)-1]
		regex, err = regexp.Compile(regexPattern)
		if err != nil {
			return nil, fmt.Errorf("invalid regex pattern: %v", err)
		}
	}

	// 搜索所有索引
	for _, index := range s.indexes {
		// 搜索函数
		if req.SymbolType == "function" || req.SymbolType == "all" || req.SymbolType == "" {
			for _, function := range index.Functions {
				if s.matchSymbol(function.Name, pattern, regex) {
					results = append(results, &SearchSymbolResponse{
						SymbolID:       function.ID,
						SymbolName:     function.Name,
						SymbolType:     "function",
						FilePath:       function.FilePath,
						LineNumber:     function.StartLine,
						RelevanceScore: s.calculateSymbolRelevance(function.Name, pattern),
						Context:        function.Signature,
					})
				}
			}
		}

		// 搜索类
		if req.SymbolType == "class" || req.SymbolType == "all" || req.SymbolType == "" {
			for _, class := range index.Classes {
				if s.matchSymbol(class.Name, pattern, regex) {
					results = append(results, &SearchSymbolResponse{
						SymbolID:       class.ID,
						SymbolName:     class.Name,
						SymbolType:     "class",
						FilePath:       class.FilePath,
						LineNumber:     class.StartLine,
						RelevanceScore: s.calculateSymbolRelevance(class.Name, pattern),
						Context:        fmt.Sprintf("class with %d methods", len(class.Methods)),
					})
				}
			}
		}

		// 搜索变量
		if req.SymbolType == "variable" || req.SymbolType == "all" || req.SymbolType == "" {
			for _, variable := range index.Variables {
				if s.matchSymbol(variable.Name, pattern, regex) {
					results = append(results, &SearchSymbolResponse{
						SymbolID:       variable.ID,
						SymbolName:     variable.Name,
						SymbolType:     "variable",
						FilePath:       variable.FilePath,
						LineNumber:     variable.Line,
						RelevanceScore: s.calculateSymbolRelevance(variable.Name, pattern),
						Context:        fmt.Sprintf("%s %s", variable.Type, variable.Scope),
					})
				}
			}
		}
	}

	// 按相关性排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].RelevanceScore > results[j].RelevanceScore
	})

	// 限制结果数量
	if req.MaxResults > 0 && len(results) > req.MaxResults {
		results = results[:req.MaxResults]
	}

	return results, nil
}

// 辅助方法

// matchFunction 匹配函数
func (s *IndexerService) matchFunction(function *Function, name string, exactMatch bool) bool {
	if exactMatch {
		return function.Name == name
	}
	return strings.Contains(strings.ToLower(function.Name), strings.ToLower(name))
}

// matchClass 匹配类
func (s *IndexerService) matchClass(class *Class, name string, exactMatch bool) bool {
	if exactMatch {
		return class.Name == name
	}
	return strings.Contains(strings.ToLower(class.Name), strings.ToLower(name))
}

// matchSymbol 匹配符号
func (s *IndexerService) matchSymbol(symbolName, pattern string, regex *regexp.Regexp) bool {
	if regex != nil {
		return regex.MatchString(symbolName)
	}
	return strings.Contains(strings.ToLower(symbolName), strings.ToLower(pattern))
}

// calculateFunctionRelevance 计算函数相关性
func (s *IndexerService) calculateFunctionRelevance(function *Function, query string) float64 {
	score := 0.0
	queryLower := strings.ToLower(query)
	nameLower := strings.ToLower(function.Name)

	// 精确匹配得分最高
	if nameLower == queryLower {
		score += 1.0
	} else if strings.HasPrefix(nameLower, queryLower) {
		score += 0.8
	} else if strings.HasSuffix(nameLower, queryLower) {
		score += 0.6
	} else if strings.Contains(nameLower, queryLower) {
		score += 0.4
	}

	// 根据函数复杂度调整分数
	if function.CyclomaticComplexity > 10 {
		score += 0.1 // 复杂函数可能更重要
	}

	return score
}

// calculateClassRelevance 计算类相关性
func (s *IndexerService) calculateClassRelevance(class *Class, query string) float64 {
	score := 0.0
	queryLower := strings.ToLower(query)
	nameLower := strings.ToLower(class.Name)

	// 精确匹配得分最高
	if nameLower == queryLower {
		score += 1.0
	} else if strings.HasPrefix(nameLower, queryLower) {
		score += 0.8
	} else if strings.HasSuffix(nameLower, queryLower) {
		score += 0.6
	} else if strings.Contains(nameLower, queryLower) {
		score += 0.4
	}

	// 根据类的方法数量调整分数
	if len(class.Methods) > 5 {
		score += 0.1 // 方法多的类可能更重要
	}

	return score
}

// calculateSymbolRelevance 计算符号相关性
func (s *IndexerService) calculateSymbolRelevance(symbolName, pattern string) float64 {
	score := 0.0
	patternLower := strings.ToLower(pattern)
	nameLower := strings.ToLower(symbolName)

	// 精确匹配得分最高
	if nameLower == patternLower {
		score = 1.0
	} else if strings.HasPrefix(nameLower, patternLower) {
		score = 0.8
	} else if strings.HasSuffix(nameLower, patternLower) {
		score = 0.6
	} else if strings.Contains(nameLower, patternLower) {
		score = 0.4
	}

	// 根据名称长度调整分数（较短的名称可能更相关）
	if len(symbolName) < 10 {
		score += 0.1
	}

	return score
}

// findCallersRecursive 递归查找调用者
func (s *IndexerService) findCallersRecursive(functionID string, maxDepth, currentDepth int, visited map[string]bool, callers *[]*CallInfo) {
	if currentDepth >= maxDepth || visited[functionID] {
		return
	}

	visited[functionID] = true

	// 查找直接调用者
	for _, index := range s.indexes {
		for _, function := range index.Functions {
			if contains(function.Calls, functionID) {
				// 找到调用者
				targetFunc := s.findFunctionByID(functionID)
				if targetFunc != nil {
					*callers = append(*callers, &CallInfo{
						CallerID:   function.ID,
						CallerName: function.Name,
						CalleeID:   functionID,
						CalleeName: targetFunc.Name,
						CallLine:   function.StartLine, // 简化实现
						CallType:   "direct",
						FilePath:   function.FilePath,
					})

					// 递归查找更深层的调用者
					if currentDepth+1 < maxDepth {
						s.findCallersRecursive(function.ID, maxDepth, currentDepth+1, visited, callers)
					}
				}
			}
		}
	}
}

// findCalleesRecursive 递归查找被调用者
func (s *IndexerService) findCalleesRecursive(functionID string, maxDepth, currentDepth int, visited map[string]bool, callees *[]*CallInfo) {
	if currentDepth >= maxDepth || visited[functionID] {
		return
	}

	visited[functionID] = true

	// 查找直接被调用者
	sourceFunc := s.findFunctionByID(functionID)
	if sourceFunc != nil {
		for _, calleeID := range sourceFunc.Calls {
			targetFunc := s.findFunctionByID(calleeID)
			if targetFunc != nil {
				*callees = append(*callees, &CallInfo{
					CallerID:   functionID,
					CallerName: sourceFunc.Name,
					CalleeID:   calleeID,
					CalleeName: targetFunc.Name,
					CallLine:   sourceFunc.StartLine, // 简化实现
					CallType:   "direct",
					FilePath:   targetFunc.FilePath,
				})

				// 递归查找更深层的被调用者
				if currentDepth+1 < maxDepth {
					s.findCalleesRecursive(calleeID, maxDepth, currentDepth+1, visited, callees)
				}
			}
		}
	}
}

// findFunctionByID 根据ID查找函数
func (s *IndexerService) findFunctionByID(functionID string) *Function {
	for _, index := range s.indexes {
		if function, exists := index.Functions[functionID]; exists {
			return function
		}
	}
	return nil
}

// 数据库查询方法

// queryFunctionFromDB 从数据库查询函数
func (s *IndexerService) queryFunctionFromDB(req *QueryFunctionRequest) ([]*Function, error) {
	var functions []*Function

	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte("index:")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var index FileIndex
				if err := json.Unmarshal(val, &index); err != nil {
					return err
				}

				// 如果指定了文件路径，只在该文件中查询
				if req.FilePath != "" && index.FilePath != req.FilePath {
					return nil
				}

				for _, function := range index.Functions {
					if s.matchFunction(function, req.FunctionName, req.ExactMatch) {
						functions = append(functions, function)
					}
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	return functions, err
}

// queryClassFromDB 从数据库查询类
func (s *IndexerService) queryClassFromDB(req *QueryClassRequest) ([]*Class, error) {
	var classes []*Class

	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte("index:")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var index FileIndex
				if err := json.Unmarshal(val, &index); err != nil {
					return err
				}

				// 如果指定了文件路径，只在该文件中查询
				if req.FilePath != "" && index.FilePath != req.FilePath {
					return nil
				}

				for _, class := range index.Classes {
					if s.matchClass(class, req.ClassName, req.ExactMatch) {
						classes = append(classes, class)
					}
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	return classes, err
}

// 辅助函数
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}