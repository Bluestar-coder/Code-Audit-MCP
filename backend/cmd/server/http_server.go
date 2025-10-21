package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"io"

	"code-audit-mcp/internal/rules"
	pb "code-audit-mcp/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// startHTTPServer 启动一个简单的 HTTP 网关，暴露 JSON API 给前端
func startHTTPServer(vuln *rules.Service, httpPort int, grpcPort int) {
	mux := http.NewServeMux()

	// 建立到 gRPC TaintAnalyzer 的连接
	grpcAddr := fmt.Sprintf("localhost:%d", grpcPort)
	conn, err := grpc.Dial(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("⚠️ Failed to dial gRPC at %s: %v", grpcAddr, err)
	}
	var taintClient pb.TaintAnalyzerClient
	if err == nil {
		taintClient = pb.NewTaintAnalyzerClient(conn)
		defer conn.Close()
	}

	// 健康检查
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// 扫描单个文件
	mux.HandleFunc("/api/scan", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			writeCORSHeaders(w)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != http.MethodPost {
			writeCORSHeaders(w)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		writeCORSHeaders(w)
		w.Header().Set("Content-Type", "application/json")

		var req pb.ScanFileRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		resp, err := vuln.ScanFile(context.Background(), &req)
		if err != nil {
			http.Error(w, fmt.Sprintf("scan error: %v", err), http.StatusInternalServerError)
			return
		}

		enc := json.NewEncoder(w)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(resp); err != nil {
			log.Printf("encode response error: %v", err)
		}
	})

	// 获取规则列表，可选 language 参数
	mux.HandleFunc("/api/rules", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			writeCORSHeaders(w)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != http.MethodGet {
			writeCORSHeaders(w)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		writeCORSHeaders(w)
		w.Header().Set("Content-Type", "application/json")

		language := r.URL.Query().Get("language")
		resp, err := vuln.GetRules(context.Background(), &pb.GetRulesRequest{Language: language})
		if err != nil {
			http.Error(w, fmt.Sprintf("rules error: %v", err), http.StatusInternalServerError)
			return
		}

		enc := json.NewEncoder(w)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(resp); err != nil {
			log.Printf("encode response error: %v", err)
		}
	})

	// Dashboard 和统计占位端点
	mux.HandleFunc("/api/dashboard", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			writeCORSHeaders(w)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != http.MethodGet {
			writeCORSHeaders(w)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeCORSHeaders(w)
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"project_stats": map[string]interface{}{
				"total_files": 0,
				"total_lines": 0,
				"total_functions": 0,
				"total_classes": 0,
				"languages": map[string]int{},
				"last_scan_time": "",
			},
			"vulnerability_stats": map[string]interface{}{
				"total": 0, "critical": 0, "high": 0, "medium": 0, "low": 0, "fixed": 0,
				"by_category": map[string]int{},
			},
			"scan_history": []map[string]interface{}{},
			"trend_data": []map[string]interface{}{},
		}
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/api/stats/project", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			writeCORSHeaders(w)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != http.MethodGet {
			writeCORSHeaders(w)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeCORSHeaders(w)
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"total_files": 0,
			"total_lines": 0,
			"total_functions": 0,
			"total_classes": 0,
			"languages": map[string]int{},
			"last_scan_time": "",
		}
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/api/stats/vulnerabilities", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			writeCORSHeaders(w)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != http.MethodGet {
			writeCORSHeaders(w)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeCORSHeaders(w)
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"total": 0, "critical": 0, "high": 0, "medium": 0, "low": 0, "fixed": 0,
			"by_category": map[string]int{},
		}
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/api/scans/history", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			writeCORSHeaders(w)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != http.MethodGet {
			writeCORSHeaders(w)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeCORSHeaders(w)
		w.Header().Set("Content-Type", "application/json")
		resp := []map[string]interface{}{}
		json.NewEncoder(w).Encode(resp)
	})

	// ===== 污点分析相关端点 =====
	// 查询污点源
	mux.HandleFunc("/api/taint/sources", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			writeCORSHeaders(w)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != http.MethodGet {
			writeCORSHeaders(w)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeCORSHeaders(w)
		w.Header().Set("Content-Type", "application/json")
		if taintClient == nil {
			http.Error(w, "taint analyzer not available", http.StatusServiceUnavailable)
			return
		}
		pattern := r.URL.Query().Get("pattern")
		language := r.URL.Query().Get("language")
		resp, err := taintClient.QuerySources(context.Background(), &pb.QuerySourcesRequest{Pattern: pattern, Language: language})
		if err != nil {
			http.Error(w, fmt.Sprintf("query sources error: %v", err), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(resp)
	})

	// 查询污点汇
	mux.HandleFunc("/api/taint/sinks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			writeCORSHeaders(w)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != http.MethodGet {
			writeCORSHeaders(w)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeCORSHeaders(w)
		w.Header().Set("Content-Type", "application/json")
		if taintClient == nil {
			http.Error(w, "taint analyzer not available", http.StatusServiceUnavailable)
			return
		}
		pattern := r.URL.Query().Get("pattern")
		language := r.URL.Query().Get("language")
		resp, err := taintClient.QuerySinks(context.Background(), &pb.QuerySinksRequest{Pattern: pattern, Language: language})
		if err != nil {
			http.Error(w, fmt.Sprintf("query sinks error: %v", err), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(resp)
	})

	// 追踪污点路径（聚合流式响应）
	mux.HandleFunc("/api/taint/trace", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			writeCORSHeaders(w)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != http.MethodPost {
			writeCORSHeaders(w)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeCORSHeaders(w)
		w.Header().Set("Content-Type", "application/json")
		if taintClient == nil {
			http.Error(w, "taint analyzer not available", http.StatusServiceUnavailable)
			return
		}
		var input struct {
			SourceFunction string `json:"source_function"`
			SinkFunction   string `json:"sink_function"`
			MaxPaths       int32  `json:"max_paths"`
		}
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&input); err != nil {
			http.Error(w, fmt.Sprintf("invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		stream, err := taintClient.TracePath(context.Background(), &pb.TracePathRequest{
			SourceFunction: input.SourceFunction,
			SinkFunction:   input.SinkFunction,
			MaxPaths:       input.MaxPaths,
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("trace path error: %v", err), http.StatusInternalServerError)
			return
		}

		// 聚合所有 PathSegment 并转换为更友好的 JSON
		var segments []struct {
			PathIndex   int32 `json:"path_index"`
			Nodes       []struct {
				NodeId       string `json:"node_id"`
				FunctionName string `json:"function_name"`
				FilePath     string `json:"file_path"`
				LineNumber   int32  `json:"line_number"`
				Operation    string `json:"operation"`
				VariableName string `json:"variable_name"`
				DataFlow     string `json:"data_flow"`
			} `json:"nodes"`
			HasSanitizer bool `json:"has_sanitizer"`
		}

		for {
			seg, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				http.Error(w, fmt.Sprintf("stream recv error: %v", err), http.StatusInternalServerError)
				return
			}
			var nodes []struct {
				NodeId       string `json:"node_id"`
				FunctionName string `json:"function_name"`
				FilePath     string `json:"file_path"`
				LineNumber   int32  `json:"line_number"`
				Operation    string `json:"operation"`
				VariableName string `json:"variable_name"`
				DataFlow     string `json:"data_flow"`
			}
			for _, n := range seg.Nodes {
				nodes = append(nodes, struct {
					NodeId       string `json:"node_id"`
					FunctionName string `json:"function_name"`
					FilePath     string `json:"file_path"`
					LineNumber   int32  `json:"line_number"`
					Operation    string `json:"operation"`
					VariableName string `json:"variable_name"`
					DataFlow     string `json:"data_flow"`
				}{
					NodeId:       n.NodeId,
					FunctionName: n.FunctionName,
					FilePath:     n.FilePath,
					LineNumber:   n.LineNumber,
					Operation:    n.Operation,
					VariableName: n.VariableName,
					DataFlow:     n.DataFlow,
				})
			}
			segments = append(segments, struct {
				PathIndex   int32 `json:"path_index"`
				Nodes       []struct {
					NodeId       string `json:"node_id"`
					FunctionName string `json:"function_name"`
					FilePath     string `json:"file_path"`
					LineNumber   int32  `json:"line_number"`
					Operation    string `json:"operation"`
					VariableName string `json:"variable_name"`
					DataFlow     string `json:"data_flow"`
				} `json:"nodes"`
				HasSanitizer bool `json:"has_sanitizer"`
			}{
				PathIndex:   seg.PathIndex,
				Nodes:       nodes,
				HasSanitizer: seg.HasSanitizer,
			})
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"paths": segments,
		})
	})

	addr := fmt.Sprintf(":%d", httpPort)
	log.Printf("🌐 HTTP API listening on %s", addr)
	if err := http.ListenAndServe(addr, withCORS(mux)); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}

// 简单 CORS 处理，允许本地前端开发访问
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeCORSHeaders(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}