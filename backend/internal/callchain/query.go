package callchain

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/dgraph-io/badger/v3"
	pb "code-audit-mcp/proto"
)

// QueryCallPath 查询调用路径
func (ccs *CallChainService) QueryCallPath(ctx context.Context, req *pb.QueryCallPathRequest) (*pb.QueryCallPathResponse, error) {
	ccs.mutex.RLock()
	defer ccs.mutex.RUnlock()
	
	var allPaths []*CallPath
	
	// 在所有调用图中查找路径
	for _, graph := range ccs.callGraphs {
		paths := ccs.findPathsBetweenFunctions(graph, req.SourceFunction, req.TargetFunction, 10)
		allPaths = append(allPaths, paths...)
	}
	
	// 按权重排序
	sort.Slice(allPaths, func(i, j int) bool {
		return allPaths[i].Weight > allPaths[j].Weight
	})
	
	// 限制返回的路径数量
	maxPaths := int(req.MaxPaths)
	if maxPaths == 0 || maxPaths > len(allPaths) {
		maxPaths = len(allPaths)
	}
	
	// 转换为protobuf格式
	var pbPaths []*pb.CallPath
	for i := 0; i < maxPaths; i++ {
		path := allPaths[i]
		var pbNodes []*pb.CallGraphNode
		
		for _, node := range path.Nodes {
			pbNode := &pb.CallGraphNode{
				NodeId:       node.ID,
				FunctionName: node.FunctionName,
				FilePath:     node.FilePath,
				LineNumber:   node.LineNumber,
				InDegree:     node.InDegree,
				OutDegree:    node.OutDegree,
				IsExternal:   node.IsExternal,
				IsRecursive:  node.IsRecursive,
				NodeType:     node.NodeType,
			}
			pbNodes = append(pbNodes, pbNode)
		}
		
		pbPath := &pb.CallPath{
			PathIndex:  path.PathIndex,
			Nodes:      pbNodes,
			PathLength: path.PathLength,
		}
		pbPaths = append(pbPaths, pbPath)
	}
	
	return &pb.QueryCallPathResponse{
		Paths:      pbPaths,
		TotalPaths: int32(len(allPaths)),
		PathExists: len(allPaths) > 0,
	}, nil
}

// findPathsBetweenFunctions 查找函数间的路径
func (ccs *CallChainService) findPathsBetweenFunctions(graph *CallGraph, sourceFunc, targetFunc string, maxDepth int) []*CallPath {
	var paths []*CallPath
	
	// 查找源函数节点
	var sourceNode *CallGraphNode
	for _, node := range graph.Nodes {
		if node.FunctionName == sourceFunc {
			sourceNode = node
			break
		}
	}
	
	if sourceNode == nil {
		return paths
	}
	
	// 查找目标函数节点
	var targetNode *CallGraphNode
	for _, node := range graph.Nodes {
		if node.FunctionName == targetFunc {
			targetNode = node
			break
		}
	}
	
	if targetNode == nil {
		return paths
	}
	
	// 使用DFS查找路径
	visited := make(map[string]bool)
	currentPath := []*CallGraphNode{sourceNode}
	ccs.dfsPathSearch(graph, sourceNode, targetNode, currentPath, visited, &paths, maxDepth, 0)
	
	return paths
}

// dfsPathSearch 深度优先搜索路径
func (ccs *CallChainService) dfsPathSearch(graph *CallGraph, current, target *CallGraphNode, 
	currentPath []*CallGraphNode, visited map[string]bool, paths *[]*CallPath, maxDepth, depth int) {
	
	if depth > maxDepth {
		return
	}
	
	if current.ID == target.ID {
		// 找到路径
		pathCopy := make([]*CallGraphNode, len(currentPath))
		copy(pathCopy, currentPath)
		
		path := &CallPath{
			PathIndex:  int32(len(*paths) + 1),
			Nodes:      pathCopy,
			PathLength: int32(len(pathCopy)),
			Weight:     ccs.calculatePathWeight(pathCopy),
		}
		*paths = append(*paths, path)
		return
	}
	
	visited[current.ID] = true
	defer delete(visited, current.ID)
	
	// 遍历当前节点的所有出边
	if edges, exists := graph.Edges[current.ID]; exists {
		for _, edge := range edges {
			if !visited[edge.TargetID] {
				if nextNode, exists := graph.Nodes[edge.TargetID]; exists {
					newPath := append(currentPath, nextNode)
					ccs.dfsPathSearch(graph, nextNode, target, newPath, visited, paths, maxDepth, depth+1)
				}
			}
		}
	}
}

// calculatePathWeight 计算路径权重
func (ccs *CallChainService) calculatePathWeight(nodes []*CallGraphNode) float64 {
	if len(nodes) == 0 {
		return 0.0
	}
	
	weight := 1.0
	
	// 路径越短权重越高
	weight = weight / float64(len(nodes))
	
	// 外部函数降低权重
	for _, node := range nodes {
		if node.IsExternal {
			weight *= 0.8
		}
		if node.IsRecursive {
			weight *= 0.9
		}
	}
	
	return weight
}

// QueryCallDepth 查询调用深度
func (ccs *CallChainService) QueryCallDepth(ctx context.Context, req *pb.QueryCallDepthRequest) (*pb.QueryCallDepthResponse, error) {
	ccs.mutex.RLock()
	defer ccs.mutex.RUnlock()
	
	var targetNode *CallGraphNode
	var graph *CallGraph
	
	// 查找目标函数
	for _, g := range ccs.callGraphs {
		for _, node := range g.Nodes {
			if node.FunctionName == req.FunctionName {
				targetNode = node
				graph = g
				break
			}
		}
		if targetNode != nil {
			break
		}
	}
	
	if targetNode == nil {
		return &pb.QueryCallDepthResponse{
			IncomingDepth: 0,
			OutgoingDepth: 0,
			MaxDepth:      0,
			Levels:        []*pb.DepthLevel{},
		}, nil
	}
	
	var incomingDepth, outgoingDepth int32
	var levels []*pb.DepthLevel
	
	if req.Direction == "incoming" || req.Direction == "both" {
		incomingLevels := ccs.calculateIncomingDepth(graph, targetNode, 10)
		incomingDepth = int32(len(incomingLevels))
		levels = append(levels, incomingLevels...)
	}
	
	if req.Direction == "outgoing" || req.Direction == "both" {
		outgoingLevels := ccs.calculateOutgoingDepth(graph, targetNode, 10)
		outgoingDepth = int32(len(outgoingLevels))
		levels = append(levels, outgoingLevels...)
	}
	
	maxDepth := incomingDepth
	if outgoingDepth > maxDepth {
		maxDepth = outgoingDepth
	}
	
	return &pb.QueryCallDepthResponse{
		IncomingDepth: incomingDepth,
		OutgoingDepth: outgoingDepth,
		MaxDepth:      maxDepth,
		Levels:        levels,
	}, nil
}


// calculateIncomingDepth 计算入向深度
func (ccs *CallChainService) calculateIncomingDepth(graph *CallGraph, target *CallGraphNode, maxDepth int) []*pb.DepthLevel {
	var levels []*pb.DepthLevel
	visited := make(map[string]bool)
	currentLevel := []string{target.ID}
	
	for depth := 0; depth < maxDepth && len(currentLevel) > 0; depth++ {
		var nextLevel []string
		var functions []string
		
		for _, nodeID := range currentLevel {
			if visited[nodeID] {
				continue
			}
			visited[nodeID] = true
			
			if node, exists := graph.Nodes[nodeID]; exists {
				functions = append(functions, node.FunctionName)
			}
			
			// 查找调用当前节点的函数
			for sourceID, edges := range graph.Edges {
				for _, edge := range edges {
					if edge.TargetID == nodeID && !visited[sourceID] {
						nextLevel = append(nextLevel, sourceID)
					}
				}
			}
		}
		
		if len(functions) > 0 {
			level := &pb.DepthLevel{
				Level:     int32(depth + 1),
				Functions: functions,
			}
			levels = append(levels, level)
		}
		
		currentLevel = nextLevel
	}
	
	return levels
}

// calculateOutgoingDepth 计算出向深度
func (ccs *CallChainService) calculateOutgoingDepth(graph *CallGraph, source *CallGraphNode, maxDepth int) []*pb.DepthLevel {
	var levels []*pb.DepthLevel
	visited := make(map[string]bool)
	currentLevel := []string{source.ID}
	
	for depth := 0; depth < maxDepth && len(currentLevel) > 0; depth++ {
		var nextLevel []string
		var functions []string
		
		for _, nodeID := range currentLevel {
			if visited[nodeID] {
				continue
			}
			visited[nodeID] = true
			
			if node, exists := graph.Nodes[nodeID]; exists {
				functions = append(functions, node.FunctionName)
			}
			
			// 查找当前节点调用的函数
			if edges, exists := graph.Edges[nodeID]; exists {
				for _, edge := range edges {
					if !visited[edge.TargetID] {
						nextLevel = append(nextLevel, edge.TargetID)
					}
				}
			}
		}
		
		if len(functions) > 0 {
			level := &pb.DepthLevel{
				Level:     int32(depth + 1),
				Functions: functions,
			}
			levels = append(levels, level)
		}
		
		currentLevel = nextLevel
	}
	
	return levels
}

// AnalyzeCycles 分析调用循环
func (ccs *CallChainService) AnalyzeCycles(ctx context.Context, req *pb.AnalyzeCyclesRequest) (*pb.AnalyzeCyclesResponse, error) {
	ccs.mutex.RLock()
	defer ccs.mutex.RUnlock()
	
	var allCycles []*CallCycle
	
	// 在所有调用图中查找循环
	for _, graph := range ccs.callGraphs {
		cycles := ccs.detectCycles(graph, req.StartFunction)
		allCycles = append(allCycles, cycles...)
	}
	
	// 转换为protobuf格式
	var pbCycles []*pb.CallCycle
	for i, cycle := range allCycles {
		pbCycle := &pb.CallCycle{
			CycleId:     cycle.CycleID,
			Functions:   cycle.Functions,
			CycleLength: cycle.CycleLength,
			Lines:       cycle.Lines,
		}
		pbCycles = append(pbCycles, pbCycle)
		
		if i >= 100 { // 限制返回的循环数量
			break
		}
	}
	
	return &pb.AnalyzeCyclesResponse{
		HasCycles:   len(allCycles) > 0,
		Cycles:      pbCycles,
		TotalCycles: int32(len(allCycles)),
	}, nil
}

// detectCycles 检测调用循环
func (ccs *CallChainService) detectCycles(graph *CallGraph, startFunction string) []*CallCycle {
	var cycles []*CallCycle
	
	// 查找起始函数节点
	var startNode *CallGraphNode
	for _, node := range graph.Nodes {
		if startFunction == "" || node.FunctionName == startFunction {
			startNode = node
			break
		}
	}
	
	if startNode == nil {
		return cycles
	}
	
	// 使用DFS检测循环
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)
	path := []string{}
	
	ccs.dfsCycleDetection(graph, startNode.ID, visited, recursionStack, path, &cycles)
	
	return cycles
}

// dfsCycleDetection DFS循环检测
func (ccs *CallChainService) dfsCycleDetection(graph *CallGraph, nodeID string, 
	visited, recursionStack map[string]bool, path []string, cycles *[]*CallCycle) {
	
	visited[nodeID] = true
	recursionStack[nodeID] = true
	path = append(path, nodeID)
	
	// 遍历所有邻接节点
	if edges, exists := graph.Edges[nodeID]; exists {
		for _, edge := range edges {
			targetID := edge.TargetID
			
			if !visited[targetID] {
				ccs.dfsCycleDetection(graph, targetID, visited, recursionStack, path, cycles)
			} else if recursionStack[targetID] {
				// 找到循环
				cycleStart := -1
				for i, id := range path {
					if id == targetID {
						cycleStart = i
						break
					}
				}
				
				if cycleStart >= 0 {
					cyclePath := path[cycleStart:]
					cyclePath = append(cyclePath, targetID) // 闭合循环
					
					var functions []string
					var lines []int32
					for _, id := range cyclePath {
						if node, exists := graph.Nodes[id]; exists {
							functions = append(functions, node.FunctionName)
							lines = append(lines, node.LineNumber)
						}
					}
					
					cycle := &CallCycle{
						CycleID:     int32(len(*cycles) + 1),
						Functions:   functions,
						CycleLength: int32(len(functions) - 1), // 减去重复的起始节点
						Lines:       lines,
						Weight:      1.0,
					}
					*cycles = append(*cycles, cycle)
				}
			}
		}
	}
	
	recursionStack[nodeID] = false
}

// FindDeadCode 查找死代码
func (ccs *CallChainService) FindDeadCode(ctx context.Context, req *pb.FindDeadCodeRequest) (*pb.FindDeadCodeResponse, error) {
	ccs.mutex.RLock()
	defer ccs.mutex.RUnlock()
	
	var deadFunctions []*DeadCodeInfo
	
	// 在所有调用图中查找死代码
	for _, graph := range ccs.callGraphs {
		if strings.Contains(graph.ID, req.FilePath) || req.FilePath == "" {
			dead := ccs.findDeadFunctions(graph, req.EntryPoints)
			deadFunctions = append(deadFunctions, dead...)
		}
	}
	
	// 转换为protobuf格式
	var pbDeadFunctions []*pb.DeadCodeInfo
	for _, dead := range deadFunctions {
		pbDead := &pb.DeadCodeInfo{
			FunctionName: dead.FunctionName,
			FilePath:     dead.FilePath,
			StartLine:    dead.StartLine,
			EndLine:      dead.EndLine,
			LineCount:    dead.LineCount,
			Complexity:   dead.Complexity,
			Reason:       dead.Reason,
		}
		pbDeadFunctions = append(pbDeadFunctions, pbDead)
	}
	
	return &pb.FindDeadCodeResponse{
		DeadFunctions:  pbDeadFunctions,
		TotalDeadCount: int32(len(deadFunctions)),
	}, nil
}

// findDeadFunctions 查找死函数
func (ccs *CallChainService) findDeadFunctions(graph *CallGraph, entryPoints []string) []*DeadCodeInfo {
	var deadFunctions []*DeadCodeInfo
	
	// 如果没有指定入口点，使用图中的入口点
	if len(entryPoints) == 0 {
		entryPoints = graph.EntryPoints
	}
	
	// 如果仍然没有入口点，使用所有入度为0的函数
	if len(entryPoints) == 0 {
		for _, node := range graph.Nodes {
			if node.InDegree == 0 && !node.IsExternal {
				entryPoints = append(entryPoints, node.FunctionName)
			}
		}
	}
	
	// 从入口点开始标记可达的函数
	reachable := make(map[string]bool)
	for _, entryPoint := range entryPoints {
		ccs.markReachableFunctions(graph, entryPoint, reachable)
	}
	
	// 查找不可达的函数
	for _, node := range graph.Nodes {
		if !node.IsExternal && !reachable[node.ID] {
			dead := &DeadCodeInfo{
				FunctionName: node.FunctionName,
				FilePath:     node.FilePath,
				StartLine:    node.LineNumber,
				EndLine:      node.LineNumber + 10, // 估算结束行
				LineCount:    10,                   // 估算行数
				Complexity:   node.Complexity,
				Reason:       "Not reachable from any entry point",
			}
			deadFunctions = append(deadFunctions, dead)
		}
	}
	
	return deadFunctions
}

// markReachableFunctions 标记可达函数
func (ccs *CallChainService) markReachableFunctions(graph *CallGraph, functionName string, reachable map[string]bool) {
	// 查找函数节点
	var startNode *CallGraphNode
	for _, node := range graph.Nodes {
		if node.FunctionName == functionName {
			startNode = node
			break
		}
	}
	
	if startNode == nil || reachable[startNode.ID] {
		return
	}
	
	// 标记当前函数为可达
	reachable[startNode.ID] = true
	
	// 递归标记所有被调用的函数
	if edges, exists := graph.Edges[startNode.ID]; exists {
		for _, edge := range edges {
			if targetNode, exists := graph.Nodes[edge.TargetID]; exists {
				ccs.markReachableFunctions(graph, targetNode.FunctionName, reachable)
			}
		}
	}
}

// loadCallGraphFromDB 从数据库加载调用图
func (ccs *CallChainService) loadCallGraphFromDB(graphID string) (*CallGraph, error) {
	graph := &CallGraph{
		ID:    graphID,
		Nodes: make(map[string]*CallGraphNode),
		Edges: make(map[string][]*CallGraphEdge),
	}
	
	err := ccs.db.View(func(txn *badger.Txn) error {
		// 加载节点
		nodePrefix := fmt.Sprintf("node:%s:", graphID)
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte(nodePrefix)
		
		it := txn.NewIterator(opts)
		defer it.Close()
		
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				parts := strings.Split(string(val), "|")
				if len(parts) >= 9 {
					node := &CallGraphNode{
						ID:           parts[0],
						FunctionName: parts[1],
						FilePath:     parts[2],
						NodeType:     parts[8],
					}
					// 解析其他字段...
					graph.Nodes[node.ID] = node
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		
		// 加载边
		edgePrefix := fmt.Sprintf("edge:%s:", graphID)
		opts.Prefix = []byte(edgePrefix)
		
		it = txn.NewIterator(opts)
		defer it.Close()
		
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				parts := strings.Split(string(val), "|")
				if len(parts) >= 6 {
					edge := &CallGraphEdge{
						ID:       parts[0],
						SourceID: parts[1],
						TargetID: parts[2],
						CallType: parts[3],
					}
					// 解析其他字段...
					graph.Edges[edge.SourceID] = append(graph.Edges[edge.SourceID], edge)
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		
		return nil
	})
	
	return graph, err
}