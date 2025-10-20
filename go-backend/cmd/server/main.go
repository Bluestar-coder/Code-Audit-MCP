package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	grpcpkg "code-audit-mcp/pkg/grpc"
	pb "code-audit-mcp/proto"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50051, "gRPC server port")
)

func main() {
	flag.Parse()

	// 创建监听器
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// 创建gRPC服务器，配置高性能参数
	opts := []grpc.ServerOption{
		// 设置最大接收消息大小（100MB）
		grpc.MaxRecvMsgSize(100 * 1024 * 1024),
		// 设置最大发送消息大小（100MB）
		grpc.MaxSendMsgSize(100 * 1024 * 1024),
	}

	s := grpc.NewServer(opts...)

	// 创建服务实例
	log.Println("📋 Initializing gRPC services...")
	parserService := grpcpkg.NewASTParserService()
	indexerService := grpcpkg.NewIndexerService()
	taintService := grpcpkg.NewTaintAnalyzerService()
	callChainService := grpcpkg.NewCallChainAnalyzerService()

	// 注册所有服务到 gRPC 服务器
	log.Println("📡 Registering services...")
	pb.RegisterASTParserServer(s, parserService)
	log.Println("  ✅ ASTParserServer registered")

	pb.RegisterIndexerServer(s, indexerService)
	log.Println("  ✅ IndexerServer registered")

	pb.RegisterTaintAnalyzerServer(s, taintService)
	log.Println("  ✅ TaintAnalyzerServer registered")

	pb.RegisterCallChainAnalyzerServer(s, callChainService)
	log.Println("  ✅ CallChainAnalyzerServer registered")

	log.Printf("🚀 gRPC server listening on :%d\n", *port)
	log.Println("📊 Available services:")
	log.Println("   - ASTParser (2 methods)")
	log.Println("   - Indexer (6 methods)")
	log.Println("   - TaintAnalyzer (4 methods)")
	log.Println("   - CallChainAnalyzer (5 methods)")
	log.Println("   Total: 17 gRPC methods")

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
