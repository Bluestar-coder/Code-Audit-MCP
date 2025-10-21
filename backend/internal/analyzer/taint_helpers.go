package analyzer

import (
	"context"
	"fmt"
	"go/ast"
	"strings"

	"google.golang.org/grpc"
	pb "code-audit-mcp/proto"
)



// isSourceCall 检查是否为污点源调用
func (ta *TaintAnalyzer) isSourceCall(call *ast.CallExpr) bool {
	funcName := ta.extractFunctionName(call)
	
	for _, source := range ta.sources {
		if strings.Contains(funcName, source.Name) {
			return true
		}
	}
	
	return false
}

// isSourceSelector 检查是否为污点源选择器
func (ta *TaintAnalyzer) isSourceSelector(sel *ast.SelectorExpr) bool {
	selectorName := ta.extractSelectorName(sel)
	
	for _, source := range ta.sources {
		if strings.Contains(selectorName, source.Name) {
			return true
		}
	}
	
	return false
}

// isSinkCall 检查是否为污点汇调用
func (ta *TaintAnalyzer) isSinkCall(call *ast.CallExpr) bool {
	funcName := ta.extractFunctionName(call)
	
	for _, sink := range ta.sinks {
		if strings.Contains(funcName, sink.Name) {
			return true
		}
	}
	
	return false
}

// isSinkAssignment 检查是否为污点汇赋值
func (ta *TaintAnalyzer) isSinkAssignment(assign *ast.AssignStmt) bool {
	for _, lhs := range assign.Lhs {
		lhsName := ta.extractExpressionName(lhs)
		for _, sink := range ta.sinks {
			if strings.Contains(lhsName, sink.Name) {
				return true
			}
		}
	}
	
	return false
}

// isSanitizerCall 检查是否为净化函数调用
func (ta *TaintAnalyzer) isSanitizerCall(call *ast.CallExpr) bool {
	funcName := ta.extractFunctionName(call)
	
	for _, sanitizer := range ta.sanitizers {
		if strings.Contains(funcName, sanitizer.Name) {
			return true
		}
	}
	
	return false
}

// extractFunctionName 提取函数名
func (ta *TaintAnalyzer) extractFunctionName(call *ast.CallExpr) string {
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

// extractVariableName 从函数调用中提取变量名
func (ta *TaintAnalyzer) extractVariableName(call *ast.CallExpr) string {
	if len(call.Args) > 0 {
		return ta.extractExpressionName(call.Args[0])
	}
	return ""
}

// extractSelectorName 提取选择器名称
func (ta *TaintAnalyzer) extractSelectorName(sel *ast.SelectorExpr) string {
	if x, ok := sel.X.(*ast.Ident); ok {
		return x.Name + "." + sel.Sel.Name
	}
	return sel.Sel.Name
}

// extractAssignmentTarget 提取赋值目标
func (ta *TaintAnalyzer) extractAssignmentTarget(assign *ast.AssignStmt) string {
	if len(assign.Lhs) > 0 {
		return ta.extractExpressionName(assign.Lhs[0])
	}
	return ""
}

// extractExpressionName 提取表达式名称
func (ta *TaintAnalyzer) extractExpressionName(expr ast.Expr) string {
	switch x := expr.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.SelectorExpr:
		if base, ok := x.X.(*ast.Ident); ok {
			return base.Name + "." + x.Sel.Name
		}
		return x.Sel.Name
	case *ast.IndexExpr:
		if base, ok := x.X.(*ast.Ident); ok {
			return base.Name
		}
		return ta.extractExpressionName(x.X)
	case *ast.CallExpr:
		return ta.extractFunctionName(x)
	case *ast.BasicLit:
		return x.Value
	default:
		return ""
	}
}

// TracePath 追踪污点路径 (流式响应)
func (ta *TaintAnalyzer) TracePath(req *pb.TracePathRequest, stream grpc.ServerStreamingServer[pb.PathSegment]) error {
	// 查找从源函数到汇函数的路径
	paths := ta.findPathsBetweenFunctions(req.SourceFunction, req.SinkFunction, int(req.MaxPaths))
	
	// 发送路径段
	for i, path := range paths {
		// 转换内部PathNode到protobuf PathNode
		var pbNodes []*pb.PathNode
		for _, node := range path.Nodes {
			pbNode := &pb.PathNode{
				NodeId:       node.NodeID,
				FunctionName: node.FunctionName,
				FilePath:     node.FilePath,
				LineNumber:   node.LineNumber,
				Operation:    node.Operation,
				VariableName: node.VariableName,
				DataFlow:     node.DataFlow,
			}
			pbNodes = append(pbNodes, pbNode)
		}
		
		segment := &pb.PathSegment{
			PathIndex:    int32(i),
			Nodes:        pbNodes,
			HasSanitizer: path.HasSanitizer,
		}
		
		if err := stream.Send(segment); err != nil {
			return fmt.Errorf("failed to send path segment: %v", err)
		}
	}
	
	return nil
}

// findPathsBetweenFunctions 查找函数间的路径
func (ta *TaintAnalyzer) findPathsBetweenFunctions(sourceFunc, sinkFunc string, maxPaths int) []*TaintPath {
	// 这里应该实现实际的路径查找逻辑
	// 目前返回一个示例路径
	var paths []*TaintPath
	
	if maxPaths > 0 {
		examplePath := &TaintPath{
			Source: &PathNode{
				NodeID:       "source_1",
				FunctionName: sourceFunc,
				FilePath:     "example.go",
				LineNumber:   10,
				Operation:    "assignment",
				VariableName: "userInput",
				DataFlow:     "source",
			},
			Sink: &PathNode{
				NodeID:       "sink_1",
				FunctionName: sinkFunc,
				FilePath:     "example.go",
				LineNumber:   20,
				Operation:    "call",
				VariableName: "query",
				DataFlow:     "sink",
			},
			Nodes: []*PathNode{
				{
					NodeID:       "source_1",
					FunctionName: sourceFunc,
					FilePath:     "example.go",
					LineNumber:   10,
					Operation:    "assignment",
					VariableName: "userInput",
					DataFlow:     "source",
				},
				{
					NodeID:       "sink_1",
					FunctionName: sinkFunc,
					FilePath:     "example.go",
					LineNumber:   20,
					Operation:    "call",
					VariableName: "query",
					DataFlow:     "sink",
				},
			},
			HasSanitizer: false,
			Confidence:   0.8,
		}
		paths = append(paths, examplePath)
	}
	
	return paths
}

// QuerySources 查询污点源
func (ta *TaintAnalyzer) QuerySources(ctx context.Context, req *pb.QuerySourcesRequest) (*pb.QuerySourcesResponse, error) {
	var sources []*pb.SourceInfo
	
	for _, source := range ta.sources {
		if req.Pattern == "" || strings.Contains(strings.ToLower(source.Name), strings.ToLower(req.Pattern)) {
			sourceInfo := &pb.SourceInfo{
				Id:          fmt.Sprintf("source_%d", len(sources)+1),
				Name:        source.Name,
				Type:        source.Type,
				Keywords:    source.Keywords,
				Description: source.Description,
			}
			sources = append(sources, sourceInfo)
		}
	}
	
	return &pb.QuerySourcesResponse{
		Sources:    sources,
		TotalCount: int32(len(sources)),
	}, nil
}

// QuerySinks 查询污点汇
func (ta *TaintAnalyzer) QuerySinks(ctx context.Context, req *pb.QuerySinksRequest) (*pb.QuerySinksResponse, error) {
	var sinks []*pb.SinkInfo
	
	for _, sink := range ta.sinks {
		if req.Pattern == "" || strings.Contains(strings.ToLower(sink.Name), strings.ToLower(req.Pattern)) {
			sinkInfo := &pb.SinkInfo{
				Id:                 fmt.Sprintf("sink_%d", len(sinks)+1),
				Name:               sink.Name,
				Type:               sink.Type,
				Keywords:           sink.Keywords,
				VulnerabilityType:  sink.VulnerabilityType,
				Description:        sink.Description,
			}
			sinks = append(sinks, sinkInfo)
		}
	}
	
	return &pb.QuerySinksResponse{
		Sinks:      sinks,
		TotalCount: int32(len(sinks)),
	}, nil
}