package grpc

import (
	"context"
	"log"

	"code-audit-mcp/internal/indexer"
	pb "code-audit-mcp/proto"
)

// IndexerService 实现索引服务
type IndexerService struct {
	pb.UnimplementedIndexerServer
	service *indexer.IndexerService
}

// NewIndexerService 创建新的索引服务
func NewIndexerService(dbPath string) (*IndexerService, error) {
	service, err := indexer.NewIndexerService(dbPath)
	if err != nil {
		return nil, err
	}
	
	return &IndexerService{
		service: service,
	}, nil
}

// Close 关闭服务
func (s *IndexerService) Close() error {
	if s.service != nil {
		return s.service.Close()
	}
	return nil
}

// BuildIndex 构建代码索引
func (s *IndexerService) BuildIndex(ctx context.Context, req *pb.BuildIndexRequest) (*pb.BuildIndexResponse, error) {
	log.Printf("🔨 Building index for file: %s (language: %s)", req.FilePath, req.Language)
	
	result, err := s.service.BuildIndex(req.FilePath, req.Language, req.AstData, req.Incremental)
	if err != nil {
		return &pb.BuildIndexResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}
	
	return &pb.BuildIndexResponse{
		Success:           result.Success,
		IndexId:           result.IndexID,
		FunctionsIndexed:  int32(result.FunctionsIndexed),
		ClassesIndexed:    int32(result.ClassesIndexed),
		VariablesIndexed:  int32(result.VariablesIndexed),
	}, nil
}

// QueryFunction 查询函数信息
func (s *IndexerService) QueryFunction(ctx context.Context, req *pb.QueryFunctionRequest) (*pb.QueryFunctionResponse, error) {
	log.Printf("🔍 Querying function: %s", req.FunctionName)
	
	queryReq := &indexer.QueryFunctionRequest{
		FunctionName: req.FunctionName,
		FilePath:     req.FilePath,
		ExactMatch:   req.ExactMatch,
	}
	
	result, err := s.service.QueryFunction(queryReq)
	if err != nil {
		return &pb.QueryFunctionResponse{}, err
	}
	
	var functions []*pb.FunctionInfo
	for _, fn := range result.Functions {
		var params []string
		for _, param := range fn.Parameters {
			params = append(params, param.Name+" "+param.Type)
		}
		
		functions = append(functions, &pb.FunctionInfo{
			Id:                    fn.ID,
			Name:                  fn.Name,
			FilePath:              fn.FilePath,
			StartLine:             int32(fn.StartLine),
			EndLine:               int32(fn.EndLine),
			Signature:             fn.Signature,
			CyclomaticComplexity:  int32(fn.CyclomaticComplexity),
			Parameters:            params,
			ReturnType:            fn.ReturnType,
		})
	}
	
	return &pb.QueryFunctionResponse{
		Functions:   functions,
		TotalCount:  int32(len(functions)),
	}, nil
}

// QueryClass 查询类信息
func (s *IndexerService) QueryClass(ctx context.Context, req *pb.QueryClassRequest) (*pb.QueryClassResponse, error) {
	log.Printf("🔍 Querying class: %s", req.ClassName)
	
	queryReq := &indexer.QueryClassRequest{
		ClassName:  req.ClassName,
		FilePath:   req.FilePath,
		ExactMatch: req.ExactMatch,
	}
	
	result, err := s.service.QueryClass(queryReq)
	if err != nil {
		return &pb.QueryClassResponse{}, err
	}
	
	var classes []*pb.ClassInfo
	for _, cls := range result.Classes {
		var methods []*pb.MethodInfo
		for _, method := range cls.Methods {
			methods = append(methods, &pb.MethodInfo{
				Id:         method.ID,
				Name:       method.Name,
				IsStatic:   method.IsStatic,
				IsPrivate:  method.IsPrivate,
				ReturnType: method.ReturnType,
			})
		}
		
		var fields []string
		for _, field := range cls.Fields {
			fields = append(fields, field.Name+" "+field.Type)
		}
		
		classes = append(classes, &pb.ClassInfo{
			Id:         cls.ID,
			Name:       cls.Name,
			FilePath:   cls.FilePath,
			StartLine:  int32(cls.StartLine),
			EndLine:    int32(cls.EndLine),
			BaseClass:  cls.BaseClass,
			Interfaces: cls.Interfaces,
			Methods:    methods,
			Fields:     fields,
		})
	}
	
	return &pb.QueryClassResponse{
		Classes:    classes,
		TotalCount: int32(len(classes)),
	}, nil
}

// QueryCallers 查询调用者
func (s *IndexerService) QueryCallers(ctx context.Context, req *pb.QueryCallersRequest) (*pb.QueryCallersResponse, error) {
	log.Printf("☎️ Querying callers for: %s", req.FunctionId)

	queryReq := &indexer.QueryCallersRequest{
		FunctionID: req.FunctionId,
		MaxDepth:   int(req.MaxDepth),
	}

	result, err := s.service.QueryCallers(queryReq)
	if err != nil {
		return &pb.QueryCallersResponse{}, err
	}

	var callers []*pb.CallInfo
	for _, call := range result.Callers {
		callers = append(callers, &pb.CallInfo{
			CallerId:   call.CallerID,
			CallerName: call.CallerName,
			CalleeId:   call.CalleeID,
			CalleeName: call.CalleeName,
			CallLine:   int32(call.CallLine),
			CallType:   call.CallType,
		})
	}

	return &pb.QueryCallersResponse{
		Callers:    callers,
		TotalCount: int32(len(callers)),
	}, nil
}

// QueryCallees 查询被调用者
func (s *IndexerService) QueryCallees(ctx context.Context, req *pb.QueryCalleesRequest) (*pb.QueryCalleesResponse, error) {
	log.Printf("📞 Querying callees for: %s", req.FunctionId)

	queryReq := &indexer.QueryCalleesRequest{
		FunctionID: req.FunctionId,
		MaxDepth:   int(req.MaxDepth),
	}

	result, err := s.service.QueryCallees(queryReq)
	if err != nil {
		return &pb.QueryCalleesResponse{}, err
	}

	var callees []*pb.CallInfo
	for _, call := range result.Callees {
		callees = append(callees, &pb.CallInfo{
			CallerId:   call.CallerID,
			CallerName: call.CallerName,
			CalleeId:   call.CalleeID,
			CalleeName: call.CalleeName,
			CallLine:   int32(call.CallLine),
			CallType:   call.CallType,
		})
	}

	return &pb.QueryCalleesResponse{
		Callees:    callees,
		TotalCount: int32(len(callees)),
	}, nil
}

// SearchSymbol 搜索符号
func (s *IndexerService) SearchSymbol(req *pb.SearchSymbolRequest, stream pb.Indexer_SearchSymbolServer) error {
	log.Printf("🔎 Searching symbols matching: %s", req.Pattern)

	queryReq := &indexer.SearchSymbolRequest{
		Pattern:    req.Pattern,
		SymbolType: req.SymbolType,
		MaxResults: int(req.MaxResults),
	}

	symbols, err := s.service.SearchSymbol(queryReq)
	if err != nil {
		return err
	}

	for _, symbol := range symbols {
		response := &pb.SearchSymbolResponse{
			SymbolId:       symbol.SymbolID,
			SymbolName:     symbol.SymbolName,
			SymbolType:     symbol.SymbolType,
			FilePath:       symbol.FilePath,
			LineNumber:     int32(symbol.LineNumber),
			RelevanceScore: symbol.RelevanceScore,
		}

		if err := stream.Send(response); err != nil {
			log.Printf("❌ Error sending search result: %v", err)
			return err
		}
	}

	log.Printf("✅ Symbol search completed")
	return nil
}
