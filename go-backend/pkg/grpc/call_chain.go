package grpc

import (
	"context"
	"log"

	"code-audit-mcp/internal/callchain"
	pb "code-audit-mcp/proto"
)

// CallChainAnalyzerService implements pb.CallChainAnalyzerServer
type CallChainAnalyzerService struct {
	pb.UnimplementedCallChainAnalyzerServer
	service *callchain.CallChainService
}

// NewCallChainAnalyzerService creates a new call chain analyzer service
func NewCallChainAnalyzerService(dbPath string) (*CallChainAnalyzerService, error) {
	service, err := callchain.NewCallChainService(dbPath)
	if err != nil {
		return nil, err
	}
	
	return &CallChainAnalyzerService{
		service: service,
	}, nil
}

// Close closes the service
func (s *CallChainAnalyzerService) Close() error {
	if s.service != nil {
		return s.service.Close()
	}
	return nil
}

// BuildCallGraph implements pb.CallChainAnalyzer/BuildCallGraph
func (s *CallChainAnalyzerService) BuildCallGraph(ctx context.Context, req *pb.BuildCallGraphRequest) (*pb.BuildCallGraphResponse, error) {
	log.Printf("🔗 Building call graph for: %s (entry points: %v)", req.FilePath, req.EntryPoints)

	return s.service.BuildCallGraph(ctx, req)
}

// QueryCallPath implements pb.CallChainAnalyzer/QueryCallPath
func (s *CallChainAnalyzerService) QueryCallPath(ctx context.Context, req *pb.QueryCallPathRequest) (*pb.QueryCallPathResponse, error) {
	log.Printf("🔍 Querying call path from %s to %s", req.SourceFunction, req.TargetFunction)

	return s.service.QueryCallPath(ctx, req)
}

// QueryCallDepth implements pb.CallChainAnalyzer/QueryCallDepth
func (s *CallChainAnalyzerService) QueryCallDepth(ctx context.Context, req *pb.QueryCallDepthRequest) (*pb.QueryCallDepthResponse, error) {
	log.Printf("📊 Querying call depth for function: %s", req.FunctionName)

	return s.service.QueryCallDepth(ctx, req)
}

// AnalyzeCycles implements pb.CallChainAnalyzer/AnalyzeCycles
func (s *CallChainAnalyzerService) AnalyzeCycles(ctx context.Context, req *pb.AnalyzeCyclesRequest) (*pb.AnalyzeCyclesResponse, error) {
	log.Printf("🔄 Analyzing cycles starting at: %s", req.StartFunction)

	return s.service.AnalyzeCycles(ctx, req)
}

// FindDeadCode implements pb.CallChainAnalyzer/FindDeadCode
func (s *CallChainAnalyzerService) FindDeadCode(ctx context.Context, req *pb.FindDeadCodeRequest) (*pb.FindDeadCodeResponse, error) {
	log.Printf("💀 Finding dead code in file: %s", req.FilePath)

	return s.service.FindDeadCode(ctx, req)
}
