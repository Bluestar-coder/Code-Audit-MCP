# Code Audit MCP

[English](README.en.md) | [简体中文](README.md)

![Go >=1.24](https://img.shields.io/badge/Go-%3E%3D1.24-blue) ![Python >=3.11](https://img.shields.io/badge/Python-%3E%3D3.11-blue) ![MCP Server](https://img.shields.io/badge/MCP-server-green)

AI-powered code security auditing built with Python + Go, with MCP (Model Context Protocol) integration. Provides index building, call graph analysis, taint tracing, rule-driven vulnerability scanning, AI explanations & PoC prompts, and OSV vulnerability search with optional semantic matching.

## Features
- Go backend (gRPC + HTTP): AST parsing, indexing, call chain, taint analysis, rule scanning
- Python MCP server: unified tools API for MCP Hosts (e.g., Claude Desktop)
- Web UI (optional): dashboard, vulnerability list, taint path visualization, code analysis
- Optional AI & semantic search: host LLM prompt generation; txtai semantic retrieval (optional)

## Architecture
- `go-backend/`: core analysis services (AST/index/call chain/taint/rule scan), exposes gRPC and HTTP gateway
- `python-mcp/`: MCP server and tools (call Go services / generate prompts / OSV queries / semantic search)
- `web-ui/`: React + TypeScript frontend consuming the HTTP gateway
- `proto/`: protobuf definitions and generated files (Go/Python)
- `rules/`: built-in common rules (SQL injection, XSS, path traversal)
- `docs/api/`: HTTP gateway endpoints overview (e.g., `/api/scan`, `/api/rules`, `/api/taint/*`)

## Quick Start (Windows)
- Prereqs: `Python 3.11+`, `Go 1.24+` (optional `Node.js 18+` for Web UI)
- Python env (recommended virtualenv)
```
cd python-mcp
python -m venv .venv
.\.venv\Scripts\Activate.ps1
pip install -e .            # core deps
# optional: dev deps
pip install -e .[dev]
```
- Start Go backend (default gRPC `localhost:50051`, HTTP gateway `localhost:8080`)
```
cd go-backend
go mod download
go run ./cmd/server
# or build then run:
# go build -o server.exe ./cmd/server
# .\server.exe -http-port 8080 -port 50051
```
- Start Web UI (optional)
```
cd web-ui
npm install
npm start   # visit http://localhost:3000
```
- Run Python tools for quick validation
```
cd python-mcp/src
python test_scan.py        # index build
python test_call_graph.py  # call graph
python test_taint.py       # taint tracing
python test_poc.py         # PoC prompt (AI optional)
python test_vuln.py        # OSV vulnerability search
```

## MCP Server & Tools
Entry point: `python -m code_audit_mcp.server` (for MCP Hosts like Claude Desktop)

- `scan_code`
  - Input: `path` (file or directory), `language?`
  - Function: build index via Go Indexer, collect function/class/variable stats
- `analyze_call_graph`
  - Input: `path`, `entry_point?`
  - Function: build call graph via Go CallChainAnalyzer
- `trace_taint`
  - Input: `source`, `sink`
  - Function: trace dataflow from source to sink via Go TaintAnalyzer
- `explain_code`
  - Input: `code`, `language?`
  - Function: return suggested prompts for the host LLM (server does not call SDK)
- `generate_poc`
  - Input: `vulnerability_id`, `language`, `context?`
  - Function: return suggested prompts for minimal reproducible PoC (host LLM generates)
- `search_vulnerabilities`
  - Input: `package_name`, `version?`, `ecosystem?` (e.g., `Go`/`PyPI`/`npm`)
  - Function: OSV API lookup for known vulns, optional semantic matching
- `scan_vulnerabilities`
  - Input: `file_path?`, `language?`, `content?`, `rule_ids?`
  - Function: call Go service via HTTP `/api/scan`, return `findings` and stats

## Optional AI & Semantic Search
- Environment variables to enable AI (driven by host, server does not call LLM SDK)
```
$Env:ANTHROPIC_API_KEY="your-key"
$Env:CLAUDE_MODEL="claude-3-5-sonnet-latest"   # optional
```
- Semantic search: install `txtai` to enable; otherwise fall back to keyword search

## Project Layout
- `go-backend/`: AST, indexing, call chain & taint analysis (gRPC + HTTP)
- `python-mcp/`: MCP server & tools, AI/semantic search
- `proto/`: proto definitions and Python/Go generated code
- `web-ui/`: frontend (optional)
- `examples/vulnerable_code/`: sample vulnerable code for local testing

## FAQ
- gRPC connection fails: ensure Go backend is running and listening on `localhost:50051`
- HTTP gateway unavailable: ensure started with `-http-port` exposing `localhost:8080`
- OSV query timeout: check network; failures return descriptive error messages
- AI usage: server returns prompts; host LLM generates output; no API keys needed in server process
- Web UI port conflict: use `$Env:PORT=3001; npm start` to change port

## Claude Desktop Config
Add this to Claude Desktop config (`mcpServers`):
```json
{
  "mcpServers": {
    "code-audit-mcp": {
      "command": "python",
      "args": ["-m", "code_audit_mcp.server"]
    }
  }
}
```
Notes:
- `command` and `args` point to the MCP server entry point.
- Default gRPC is `localhost:50051`; customize ports at Go backend startup.
- AI is host-driven: server returns prompts; host LLM produces results.

## End-to-End Examples
### 1) Index / Call Graph / Taint
Ensure Go backend is running, then:
```
cd python-mcp/src
python test_scan.py
python test_call_graph.py
python test_taint.py
```
Sample output (truncated):
```
{"path":"E:\\Code\\CodeAuditMcp\\go-backend\\internal\\indexer\\service.go","functions":10,"classes":5,"variables":50}
{"nodes":25,"edges":40,"path":"...service.go"}
{"paths_found":1,"source":"user_input","sink":"os/exec"}
```

### 2) Vuln Search + PoC Prompt
```
python test_vuln.py
python test_poc.py
```
Notes:
- `test_vuln.py` uses OSV (e.g., `CVE-2023-29401`, `GO-2023-1737`).
- `test_poc.py` returns suggested prompts for host LLM; use these to generate PoC.

## Screenshots
- Web UI: vulnerability scan results and severity labels
![scan-result](docs/assets/ui_scan_result.png)
- MCP tool invocation (CLI)
![mcp-cli](docs/assets/cli_mcp_call.png)

## License
MIT License.