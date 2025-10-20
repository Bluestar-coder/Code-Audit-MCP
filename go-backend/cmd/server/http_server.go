package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"code-audit-mcp/internal/rules"
	pb "code-audit-mcp/proto"
)

// startHTTPServer 启动一个简单的 HTTP 网关，暴露 JSON API 给前端
func startHTTPServer(vuln *rules.Service, port int) {
	mux := http.NewServeMux()

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

	addr := fmt.Sprintf(":%d", port)
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