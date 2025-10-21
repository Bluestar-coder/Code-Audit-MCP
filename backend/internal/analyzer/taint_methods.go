package analyzer

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"strings"
	"time"

	pb "code-audit-mcp/proto"
)

// AnalyzeTaint 分析污点
func (ta *TaintAnalyzer) AnalyzeTaint(ctx context.Context, req *pb.TaintAnalysisRequest) (*pb.TaintAnalysisResponse, error) {
	startTime := time.Now()
	
	// 读取文件内容
	content, err := ioutil.ReadFile(req.FilePath)
	if err != nil {
		return &pb.TaintAnalysisResponse{
			Success:      false,
			ErrorMessage: fmt.Sprintf("Failed to read file: %v", err),
		}, nil
	}
	
	// 解析代码
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, req.FilePath, content, parser.ParseComments)
	if err != nil {
		return &pb.TaintAnalysisResponse{
			Success:      false,
			ErrorMessage: fmt.Sprintf("Failed to parse file: %v", err),
		}, nil
	}
	
	// 执行污点分析
	vulnerabilities := ta.performTaintAnalysis(fset, node, req)
	
	analysisTime := time.Since(startTime)
	
	return &pb.TaintAnalysisResponse{
		Success:         true,
		Vulnerabilities: vulnerabilities,
		AnalysisTimeMs:  int32(analysisTime.Milliseconds()),
	}, nil
}

// performTaintAnalysis 执行污点分析
func (ta *TaintAnalyzer) performTaintAnalysis(fset *token.FileSet, node ast.Node, req *pb.TaintAnalysisRequest) []*pb.TaintVulnerability {
	var vulnerabilities []*pb.TaintVulnerability
	
	// 查找污点源
	sources := ta.findTaintSources(fset, node)
	
	// 查找污点汇
	sinks := ta.findTaintSinks(fset, node)
	
	// 查找净化函数
	sanitizers := ta.findSanitizers(fset, node)
	
	// 构建数据流图
	dataFlowGraph := ta.buildDataFlowGraph(fset, node)
	
	// 追踪从源到汇的路径
	for _, source := range sources {
		for _, sink := range sinks {
			paths := ta.tracePaths(source, sink, dataFlowGraph, sanitizers, req.MaxDepth)
			
			for _, path := range paths {
				if !path.HasSanitizer && path.Confidence > 0.5 {
					vuln := ta.createVulnerability(source, sink, path)
					vulnerabilities = append(vulnerabilities, vuln)
				}
			}
		}
	}
	
	return vulnerabilities
}

// findTaintSources 查找污点源
func (ta *TaintAnalyzer) findTaintSources(fset *token.FileSet, node ast.Node) []*PathNode {
	var sources []*PathNode
	
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CallExpr:
			if ta.isSourceCall(x) {
				pos := fset.Position(x.Pos())
				source := &PathNode{
					NodeID:       fmt.Sprintf("source_%d_%d", pos.Line, pos.Column),
					FunctionName: ta.extractFunctionName(x),
					FilePath:     pos.Filename,
					LineNumber:   int32(pos.Line),
					Operation:    "source_call",
					VariableName: ta.extractVariableName(x),
					DataFlow:     "taint_source",
				}
				sources = append(sources, source)
			}
		case *ast.SelectorExpr:
			if ta.isSourceSelector(x) {
				pos := fset.Position(x.Pos())
				source := &PathNode{
					NodeID:       fmt.Sprintf("source_%d_%d", pos.Line, pos.Column),
					FunctionName: "",
					FilePath:     pos.Filename,
					LineNumber:   int32(pos.Line),
					Operation:    "source_access",
					VariableName: ta.extractSelectorName(x),
					DataFlow:     "taint_source",
				}
				sources = append(sources, source)
			}
		}
		return true
	})
	
	return sources
}

// findTaintSinks 查找污点汇
func (ta *TaintAnalyzer) findTaintSinks(fset *token.FileSet, node ast.Node) []*PathNode {
	var sinks []*PathNode
	
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CallExpr:
			if ta.isSinkCall(x) {
				pos := fset.Position(x.Pos())
				sink := &PathNode{
					NodeID:       fmt.Sprintf("sink_%d_%d", pos.Line, pos.Column),
					FunctionName: ta.extractFunctionName(x),
					FilePath:     pos.Filename,
					LineNumber:   int32(pos.Line),
					Operation:    "sink_call",
					VariableName: ta.extractVariableName(x),
					DataFlow:     "taint_sink",
				}
				sinks = append(sinks, sink)
			}
		case *ast.AssignStmt:
			if ta.isSinkAssignment(x) {
				pos := fset.Position(x.Pos())
				sink := &PathNode{
					NodeID:       fmt.Sprintf("sink_%d_%d", pos.Line, pos.Column),
					FunctionName: "",
					FilePath:     pos.Filename,
					LineNumber:   int32(pos.Line),
					Operation:    "sink_assignment",
					VariableName: ta.extractAssignmentTarget(x),
					DataFlow:     "taint_sink",
				}
				sinks = append(sinks, sink)
			}
		}
		return true
	})
	
	return sinks
}

// findSanitizers 查找净化函数
func (ta *TaintAnalyzer) findSanitizers(fset *token.FileSet, node ast.Node) []*PathNode {
	var sanitizers []*PathNode
	
	ast.Inspect(node, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if ta.isSanitizerCall(call) {
				pos := fset.Position(call.Pos())
				sanitizer := &PathNode{
					NodeID:       fmt.Sprintf("sanitizer_%d_%d", pos.Line, pos.Column),
					FunctionName: ta.extractFunctionName(call),
					FilePath:     pos.Filename,
					LineNumber:   int32(pos.Line),
					Operation:    "sanitizer_call",
					VariableName: ta.extractVariableName(call),
					DataFlow:     "sanitizer",
				}
				sanitizers = append(sanitizers, sanitizer)
			}
		}
		return true
	})
	
	return sanitizers
}

// buildDataFlowGraph 构建数据流图
func (ta *TaintAnalyzer) buildDataFlowGraph(fset *token.FileSet, node ast.Node) map[string][]*PathNode {
	dataFlow := make(map[string][]*PathNode)
	
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.AssignStmt:
			ta.processAssignment(fset, x, dataFlow)
		case *ast.CallExpr:
			ta.processCall(fset, x, dataFlow)
		}
		return true
	})
	
	return dataFlow
}

// processAssignment 处理赋值语句
func (ta *TaintAnalyzer) processAssignment(fset *token.FileSet, assign *ast.AssignStmt, dataFlow map[string][]*PathNode) {
	pos := fset.Position(assign.Pos())
	
	for i, lhs := range assign.Lhs {
		if i < len(assign.Rhs) {
			lhsName := ta.extractExpressionName(lhs)
			rhsName := ta.extractExpressionName(assign.Rhs[i])
			
			if lhsName != "" && rhsName != "" {
				node := &PathNode{
					NodeID:       fmt.Sprintf("assign_%d_%d", pos.Line, pos.Column),
					FunctionName: "",
					FilePath:     pos.Filename,
					LineNumber:   int32(pos.Line),
					Operation:    "assignment",
					VariableName: lhsName,
					DataFlow:     fmt.Sprintf("%s = %s", lhsName, rhsName),
				}
				
				dataFlow[lhsName] = append(dataFlow[lhsName], node)
				if rhsNodes, exists := dataFlow[rhsName]; exists {
					dataFlow[lhsName] = append(dataFlow[lhsName], rhsNodes...)
				}
			}
		}
	}
}

// processCall 处理函数调用
func (ta *TaintAnalyzer) processCall(fset *token.FileSet, call *ast.CallExpr, dataFlow map[string][]*PathNode) {
	pos := fset.Position(call.Pos())
	funcName := ta.extractFunctionName(call)
	
	node := &PathNode{
		NodeID:       fmt.Sprintf("call_%d_%d", pos.Line, pos.Column),
		FunctionName: funcName,
		FilePath:     pos.Filename,
		LineNumber:   int32(pos.Line),
		Operation:    "function_call",
		VariableName: "",
		DataFlow:     fmt.Sprintf("call %s", funcName),
	}
	
	// 处理函数参数的数据流
	for _, arg := range call.Args {
		argName := ta.extractExpressionName(arg)
		if argName != "" {
			if argNodes, exists := dataFlow[argName]; exists {
				dataFlow[funcName] = append(dataFlow[funcName], argNodes...)
			}
			dataFlow[funcName] = append(dataFlow[funcName], node)
		}
	}
}

// tracePaths 追踪从源到汇的路径
func (ta *TaintAnalyzer) tracePaths(source, sink *PathNode, dataFlow map[string][]*PathNode, sanitizers []*PathNode, maxDepth int32) []*TaintPath {
	var paths []*TaintPath
	
	// 简化的路径追踪实现
	// 检查源变量是否直接或间接流向汇变量
	if ta.hasDataFlow(source.VariableName, sink.VariableName, dataFlow, int(maxDepth)) {
		path := &TaintPath{
			Source:      source,
			Sink:        sink,
			Nodes:       []*PathNode{source, sink},
			HasSanitizer: ta.pathHasSanitizer(source, sink, sanitizers),
			Confidence:  0.8, // 简化的置信度计算
		}
		paths = append(paths, path)
	}
	
	return paths
}

// hasDataFlow 检查是否存在数据流
func (ta *TaintAnalyzer) hasDataFlow(sourceVar, sinkVar string, dataFlow map[string][]*PathNode, maxDepth int) bool {
	if maxDepth <= 0 {
		return false
	}
	
	if sourceVar == sinkVar {
		return true
	}
	
	// 检查直接数据流
	if nodes, exists := dataFlow[sinkVar]; exists {
		for _, node := range nodes {
			if strings.Contains(node.DataFlow, sourceVar) {
				return true
			}
		}
	}
	
	// 递归检查间接数据流
	for varName := range dataFlow {
		if varName != sourceVar && ta.hasDataFlow(sourceVar, varName, dataFlow, maxDepth-1) {
			if ta.hasDataFlow(varName, sinkVar, dataFlow, maxDepth-1) {
				return true
			}
		}
	}
	
	return false
}

// pathHasSanitizer 检查路径是否包含净化函数
func (ta *TaintAnalyzer) pathHasSanitizer(source, sink *PathNode, sanitizers []*PathNode) bool {
	// 简化实现：检查净化函数是否在源和汇之间
	for _, sanitizer := range sanitizers {
		if sanitizer.LineNumber > source.LineNumber && sanitizer.LineNumber < sink.LineNumber {
			return true
		}
	}
	return false
}

// createVulnerability 创建漏洞对象
func (ta *TaintAnalyzer) createVulnerability(source, sink *PathNode, path *TaintPath) *pb.TaintVulnerability {
	// 根据汇的类型确定漏洞类型
	vulnType := "unknown"
	severity := "medium"
	
	for _, sinkInfo := range ta.sinks {
		if strings.Contains(sink.FunctionName, sinkInfo.Name) || 
		   strings.Contains(sink.VariableName, sinkInfo.Name) {
			vulnType = sinkInfo.VulnerabilityType
			severity = "high"
			break
		}
	}
	
	// 转换路径节点
	var pathNodes []*pb.PathNode
	for _, node := range path.Nodes {
		pathNodes = append(pathNodes, &pb.PathNode{
			NodeId:       node.NodeID,
			FunctionName: node.FunctionName,
			FilePath:     node.FilePath,
			LineNumber:   node.LineNumber,
			Operation:    node.Operation,
			VariableName: node.VariableName,
			DataFlow:     node.DataFlow,
		})
	}
	
	return &pb.TaintVulnerability{
		Id:          fmt.Sprintf("taint_%s_%s", source.NodeID, sink.NodeID),
		Type:        vulnType,
		Severity:    severity,
		Source:      fmt.Sprintf("%s:%d", source.FilePath, source.LineNumber),
		Sink:        fmt.Sprintf("%s:%d", sink.FilePath, sink.LineNumber),
		Path:        pathNodes,
		Confidence:  path.Confidence,
		Description: fmt.Sprintf("Taint flow from %s to %s", source.VariableName, sink.VariableName),
	}
}