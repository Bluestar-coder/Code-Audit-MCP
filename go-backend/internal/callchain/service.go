package callchain

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v3"
	pb "code-audit-mcp/proto"
)

// CallChainService 调用链分析服务
type CallChainService struct {
	db         *badger.DB
	callGraphs map[string]*CallGraph
	mutex      sync.RWMutex
}

// CallGraph 调用图
type CallGraph struct {
	ID        string
	Nodes     map[string]*CallGraphNode
	Edges     map[string][]*CallGraphEdge
	EntryPoints []string
	BuildTime time.Time
}

// CallGraphNode 调用图节点
type CallGraphNode struct {
	ID           string
	FunctionName string
	FilePath     string
	LineNumber   int32
	InDegree     int32
	OutDegree    int32
	IsExternal   bool
	IsRecursive  bool
	NodeType     string
	Signature    string
	Complexity   float64
}

// CallGraphEdge 调用图边
type CallGraphEdge struct {
	ID         string
	SourceID   string
	TargetID   string
	CallType   string
	LineNumber int32
	Weight     float64
}

// CallPath 调用路径
type CallPath struct {
	PathIndex  int32
	Nodes      []*CallGraphNode
	PathLength int32
	Weight     float64
}

// CallCycle 调用循环
type CallCycle struct {
	CycleID     int32
	Functions   []string
	CycleLength int32
	Lines       []int32
	Weight      float64
}

// DeadCodeInfo 死代码信息
type DeadCodeInfo struct {
	FunctionName string
	FilePath     string
	StartLine    int32
	EndLine      int32
	LineCount    int32
	Complexity   float64
	Reason       string
}

// NewCallChainService 创建新的调用链服务
func NewCallChainService(dbPath string) (*CallChainService, error) {
	opts := badger.DefaultOptions(dbPath + "/callchain")
	opts.Logger = nil // 禁用日志
	
	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open BadgerDB: %v", err)
	}

	return &CallChainService{
		db:         db,
		callGraphs: make(map[string]*CallGraph),
	}, nil
}

// Close 关闭服务
func (ccs *CallChainService) Close() error {
	return ccs.db.Close()
}

// BuildCallGraph 构建调用图
func (ccs *CallChainService) BuildCallGraph(ctx context.Context, req *pb.BuildCallGraphRequest) (*pb.BuildCallGraphResponse, error) {
	startTime := time.Now()
	
	// 读取文件内容
	content, err := ioutil.ReadFile(req.FilePath)
	if err != nil {
		return &pb.BuildCallGraphResponse{
			Success:      false,
			ErrorMessage: fmt.Sprintf("Failed to read file: %v", err),
		}, nil
	}
	
	// 解析代码
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, req.FilePath, content, parser.ParseComments)
	if err != nil {
		return &pb.BuildCallGraphResponse{
			Success:      false,
			ErrorMessage: fmt.Sprintf("Failed to parse file: %v", err),
		}, nil
	}
	
	// 构建调用图
	graph := ccs.buildCallGraphFromAST(fset, node, req)
	
	// 保存调用图
	ccs.mutex.Lock()
	ccs.callGraphs[graph.ID] = graph
	ccs.mutex.Unlock()
	
	// 保存到数据库
	err = ccs.saveCallGraphToDB(graph)
	if err != nil {
		return &pb.BuildCallGraphResponse{
			Success:      false,
			ErrorMessage: fmt.Sprintf("Failed to save call graph: %v", err),
		}, nil
	}
	
	buildTime := time.Since(startTime)
	
	return &pb.BuildCallGraphResponse{
		Success:     true,
		GraphId:     graph.ID,
		NodeCount:   int32(len(graph.Nodes)),
		EdgeCount:   int32(ccs.countEdges(graph)),
		BuildTimeMs: int32(buildTime.Milliseconds()),
	}, nil
}

// buildCallGraphFromAST 从AST构建调用图
func (ccs *CallChainService) buildCallGraphFromAST(fset *token.FileSet, node ast.Node, req *pb.BuildCallGraphRequest) *CallGraph {
	graph := &CallGraph{
		ID:          fmt.Sprintf("graph_%s_%d", req.FilePath, time.Now().Unix()),
		Nodes:       make(map[string]*CallGraphNode),
		Edges:       make(map[string][]*CallGraphEdge),
		EntryPoints: req.EntryPoints,
		BuildTime:   time.Now(),
	}
	
	// 第一遍：收集所有函数定义
	ast.Inspect(node, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			pos := fset.Position(fn.Pos())
			nodeID := fmt.Sprintf("func_%s_%d", fn.Name.Name, pos.Line)
			
			graphNode := &CallGraphNode{
				ID:           nodeID,
				FunctionName: fn.Name.Name,
				FilePath:     req.FilePath,
				LineNumber:   int32(pos.Line),
				NodeType:     "function",
				Signature:    ccs.extractFunctionSignature(fn),
				Complexity:   ccs.calculateComplexity(fn),
			}
			
			graph.Nodes[nodeID] = graphNode
		}
		return true
	})
	
	// 第二遍：收集函数调用关系
	ast.Inspect(node, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			ccs.processCallExpression(fset, call, graph, req.FilePath)
		}
		return true
	})
	
	// 计算入度和出度
	ccs.calculateDegrees(graph)
	
	// 检测递归调用
	ccs.detectRecursiveCalls(graph)
	
	return graph
}

// processCallExpression 处理函数调用表达式
func (ccs *CallChainService) processCallExpression(fset *token.FileSet, call *ast.CallExpr, graph *CallGraph, filePath string) {
	pos := fset.Position(call.Pos())
	
	// 提取被调用的函数名
	calleeName := ccs.extractCalleeName(call)
	if calleeName == "" {
		return
	}
	
	// 查找调用者函数
	callerNode := ccs.findContainingFunction(fset, call, graph)
	if callerNode == nil {
		return
	}
	
	// 查找或创建被调用函数节点
	calleeNode := ccs.findOrCreateCalleeNode(calleeName, graph, filePath)
	
	// 创建调用边
	edgeID := fmt.Sprintf("edge_%s_%s_%d", callerNode.ID, calleeNode.ID, pos.Line)
	edge := &CallGraphEdge{
		ID:         edgeID,
		SourceID:   callerNode.ID,
		TargetID:   calleeNode.ID,
		CallType:   "direct_call",
		LineNumber: int32(pos.Line),
		Weight:     1.0,
	}
	
	graph.Edges[callerNode.ID] = append(graph.Edges[callerNode.ID], edge)
}

// extractCalleeName 提取被调用函数名
func (ccs *CallChainService) extractCalleeName(call *ast.CallExpr) string {
	switch fun := call.Fun.(type) {
	case *ast.Ident:
		return fun.Name
	case *ast.SelectorExpr:
		if x, ok := fun.X.(*ast.Ident); ok {
			return x.Name + "." + fun.Sel.Name
		}
		return fun.Sel.Name
	default:
		return ""
	}
}

// findContainingFunction 查找包含调用的函数
func (ccs *CallChainService) findContainingFunction(fset *token.FileSet, call *ast.CallExpr, graph *CallGraph) *CallGraphNode {
	callPos := fset.Position(call.Pos())
	
	for _, node := range graph.Nodes {
		if node.FilePath == callPos.Filename && 
		   node.LineNumber <= int32(callPos.Line) {
			// 简化实现：假设函数按行号排序，找到最近的函数
			return node
		}
	}
	
	return nil
}

// findOrCreateCalleeNode 查找或创建被调用函数节点
func (ccs *CallChainService) findOrCreateCalleeNode(calleeName string, graph *CallGraph, filePath string) *CallGraphNode {
	// 首先在现有节点中查找
	for _, node := range graph.Nodes {
		if node.FunctionName == calleeName {
			return node
		}
	}
	
	// 如果没找到，创建外部函数节点
	nodeID := fmt.Sprintf("external_%s", calleeName)
	node := &CallGraphNode{
		ID:           nodeID,
		FunctionName: calleeName,
		FilePath:     filePath,
		LineNumber:   0,
		IsExternal:   true,
		NodeType:     "external_function",
		Signature:    calleeName + "()",
	}
	
	graph.Nodes[nodeID] = node
	return node
}

// extractFunctionSignature 提取函数签名
func (ccs *CallChainService) extractFunctionSignature(fn *ast.FuncDecl) string {
	var params []string
	if fn.Type.Params != nil {
		for _, param := range fn.Type.Params.List {
			for _, name := range param.Names {
				paramType := "interface{}"
				if param.Type != nil {
					paramType = ccs.extractTypeString(param.Type)
				}
				params = append(params, name.Name+" "+paramType)
			}
		}
	}
	
	var results []string
	if fn.Type.Results != nil {
		for _, result := range fn.Type.Results.List {
			resultType := ccs.extractTypeString(result.Type)
			results = append(results, resultType)
		}
	}
	
	signature := fn.Name.Name + "(" + strings.Join(params, ", ") + ")"
	if len(results) > 0 {
		signature += " (" + strings.Join(results, ", ") + ")"
	}
	
	return signature
}

// extractTypeString 提取类型字符串
func (ccs *CallChainService) extractTypeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		if x, ok := t.X.(*ast.Ident); ok {
			return x.Name + "." + t.Sel.Name
		}
		return t.Sel.Name
	case *ast.StarExpr:
		return "*" + ccs.extractTypeString(t.X)
	case *ast.ArrayType:
		return "[]" + ccs.extractTypeString(t.Elt)
	case *ast.MapType:
		return "map[" + ccs.extractTypeString(t.Key) + "]" + ccs.extractTypeString(t.Value)
	default:
		return "interface{}"
	}
}

// calculateComplexity 计算函数复杂度
func (ccs *CallChainService) calculateComplexity(fn *ast.FuncDecl) float64 {
	complexity := 1.0 // 基础复杂度
	
	ast.Inspect(fn, func(n ast.Node) bool {
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

// calculateDegrees 计算节点的入度和出度
func (ccs *CallChainService) calculateDegrees(graph *CallGraph) {
	// 重置度数
	for _, node := range graph.Nodes {
		node.InDegree = 0
		node.OutDegree = 0
	}
	
	// 计算出度
	for sourceID, edges := range graph.Edges {
		if node, exists := graph.Nodes[sourceID]; exists {
			node.OutDegree = int32(len(edges))
		}
	}
	
	// 计算入度
	for _, edges := range graph.Edges {
		for _, edge := range edges {
			if node, exists := graph.Nodes[edge.TargetID]; exists {
				node.InDegree++
			}
		}
	}
}

// detectRecursiveCalls 检测递归调用
func (ccs *CallChainService) detectRecursiveCalls(graph *CallGraph) {
	// 直接递归：自环
	for sourceID, edges := range graph.Edges {
		for _, edge := range edges {
			if edge.TargetID == sourceID {
				if node, exists := graph.Nodes[sourceID]; exists {
					node.IsRecursive = true
				}
			}
		}
	}

	// 间接递归：使用 Tarjan 算法检测强连通分量（SCC）
	index := 0
	indices := make(map[string]int)
	lowlink := make(map[string]int)
	onStack := make(map[string]bool)
	stack := []string{}

	var sccs [][]string

	var strongConnect func(v string)
	strongConnect = func(v string) {
		indices[v] = index
		lowlink[v] = index
		index++
		stack = append(stack, v)
		onStack[v] = true

		if edges, exists := graph.Edges[v]; exists {
			for _, edge := range edges {
				w := edge.TargetID
				if _, seen := indices[w]; !seen {
					strongConnect(w)
					if lowlink[w] < lowlink[v] {
						lowlink[v] = lowlink[w]
					}
				} else if onStack[w] && indices[w] < lowlink[v] {
					lowlink[v] = indices[w]
				}
			}
		}

		if lowlink[v] == indices[v] {
			// v 是一个 SCC 的根
			var component []string
			for {
				if len(stack) == 0 {
					break
				}
				w := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				onStack[w] = false
				component = append(component, w)
				if w == v {
					break
				}
			}
			if len(component) > 1 {
				sccs = append(sccs, component)
			}
		}
	}

	// 遍历所有节点执行强连通分量分析
	for id := range graph.Nodes {
		if _, seen := indices[id]; !seen {
			strongConnect(id)
		}
	}

	// 标记处于非单节点 SCC 的所有节点为递归
	for _, comp := range sccs {
		for _, nodeID := range comp {
			if node, exists := graph.Nodes[nodeID]; exists {
				node.IsRecursive = true
			}
		}
	}
}

// countEdges 计算边的总数
func (ccs *CallChainService) countEdges(graph *CallGraph) int {
	count := 0
	for _, edges := range graph.Edges {
		count += len(edges)
	}
	return count
}

// saveCallGraphToDB 保存调用图到数据库
func (ccs *CallChainService) saveCallGraphToDB(graph *CallGraph) error {
	return ccs.db.Update(func(txn *badger.Txn) error {
		// 保存图的基本信息
		graphKey := fmt.Sprintf("graph:%s", graph.ID)
		graphData := fmt.Sprintf("%s|%d|%d", graph.ID, len(graph.Nodes), ccs.countEdges(graph))
		err := txn.Set([]byte(graphKey), []byte(graphData))
		if err != nil {
			return err
		}
		
		// 保存节点
		for _, node := range graph.Nodes {
			nodeKey := fmt.Sprintf("node:%s:%s", graph.ID, node.ID)
			nodeData := fmt.Sprintf("%s|%s|%s|%d|%d|%d|%t|%t|%s",
				node.ID, node.FunctionName, node.FilePath, node.LineNumber,
				node.InDegree, node.OutDegree, node.IsExternal, node.IsRecursive, node.NodeType)
			err := txn.Set([]byte(nodeKey), []byte(nodeData))
			if err != nil {
				return err
			}
		}
		
		// 保存边
		for sourceID, edges := range graph.Edges {
			for _, edge := range edges {
				edgeKey := fmt.Sprintf("edge:%s:%s", graph.ID, edge.ID)
				edgeData := fmt.Sprintf("%s|%s|%s|%s|%d|%f",
					edge.ID, edge.SourceID, edge.TargetID, edge.CallType, edge.LineNumber, edge.Weight)
				err := txn.Set([]byte(edgeKey), []byte(edgeData))
				if err != nil {
					return err
				}
			}
		}
		
		return nil
	})
}