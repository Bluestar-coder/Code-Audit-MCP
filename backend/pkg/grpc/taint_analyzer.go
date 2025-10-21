package grpc

import (
	"context"
	"log"

	"code-audit-mcp/internal/analyzer"
	pb "code-audit-mcp/proto"
)

// TaintAnalyzerService implements pb.TaintAnalyzerServer
type TaintAnalyzerService struct {
	pb.UnimplementedTaintAnalyzerServer
	analyzer *analyzer.TaintAnalyzer
}

// NewTaintAnalyzerService creates a new taint analyzer service
func NewTaintAnalyzerService() *TaintAnalyzerService {
	return &TaintAnalyzerService{
		analyzer: analyzer.NewTaintAnalyzer(),
	}
}

// AnalyzeTaint implements pb.TaintAnalyzer/AnalyzeTaint
func (s *TaintAnalyzerService) AnalyzeTaint(ctx context.Context, req *pb.TaintAnalysisRequest) (*pb.TaintAnalysisResponse, error) {
	log.Printf("🔍 Analyzing taint in: %s (entry: %s)", req.FilePath, req.EntryFunction)

	// 使用实际的污点分析器
	return s.analyzer.AnalyzeTaint(ctx, req)
}

// TracePath implements pb.TaintAnalyzer/TracePath
func (s *TaintAnalyzerService) TracePath(req *pb.TracePathRequest, stream pb.TaintAnalyzer_TracePathServer) error {
	log.Printf("🔗 Tracing path from source: %s to sink: %s", req.SourceFunction, req.SinkFunction)

	// 使用实际的污点分析器
	err := s.analyzer.TracePath(req, stream)
	if err != nil {
		log.Printf("❌ Error tracing path: %v", err)
		return err
	}

	log.Printf("✅ Path tracing completed")
	return nil
}

// QuerySources implements pb.TaintAnalyzer/QuerySources
func (s *TaintAnalyzerService) QuerySources(ctx context.Context, req *pb.QuerySourcesRequest) (*pb.QuerySourcesResponse, error) {
	log.Printf("📍 Querying taint sources matching: %s", req.Pattern)

	// 使用实际的污点分析器
	return s.analyzer.QuerySources(ctx, req)
}

// QuerySinks implements pb.TaintAnalyzer/QuerySinks
func (s *TaintAnalyzerService) QuerySinks(ctx context.Context, req *pb.QuerySinksRequest) (*pb.QuerySinksResponse, error) {
	log.Printf("🎯 Querying taint sinks matching: %s", req.Pattern)

	// 使用实际的污点分析器
	return s.analyzer.QuerySinks(ctx, req)
}
