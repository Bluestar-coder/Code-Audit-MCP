# Code Audit MCP

[English](README.en.md) | [简体中文](README.md)

AI-powered code security auditing with Python + Go and MCP integration. Provides index building, call graph analysis, taint tracing, rule-based scanning, AI explanations & PoC prompts, and OSV vulnerability search.

## Features
- Backend (Go, gRPC+HTTP): AST / index / call chain / taint / rule scanning
- MCP (Python): unified tools API for MCP Hosts (e.g., Claude Desktop)
- Frontend (React, optional): dashboard and vulnerability list
- Rules (YAML): common vulnerability rules

## Architecture
- `backend/`: core analysis & service APIs (`cmd/server`, `internal`, `proto`)
- `mcp/`: MCP server & tools entry (`python -m code_audit_mcp.server`)
- `frontend/`: React app (production copied to `release/frontend`)
- `proto/`: protobuf definitions and generated code (Go/Python)
- `rules/`: YAML rule library
- `scripts/`: `check-clean-state.ps1`, `build-release.ps1`
- `release/`: distribution folder (`server.exe`, `frontend/`, `rules/common/`)

## Quick Start (Windows)
- Prereqs: `Python 3.11+`, `Go 1.24+` (Frontend needs `Node.js 18+`)
- Python env (virtualenv recommended)
```
cd mcp
python -m venv .venv
.\.venv\Scripts\Activate.ps1
pip install -e .
```
- Start backend (default gRPC `localhost:50051`, HTTP `localhost:8080`)
```
cd backend
go mod download
go run ./cmd/server
# or build and run:
# go build -o server.exe ./cmd/server
# .\server.exe -http-port 8080 -port 50051
```
- Start frontend (optional)
```
cd frontend
npm install
npm start   # visit http://localhost:3000
```

## Release Build
- One command:
```
powershell -ExecutionPolicy Bypass -File scripts\build-release.ps1
```
- Outputs: `release/server.exe`, `release/frontend/`, `release/rules/common/`
- See `release/README-release.md` for usage
- Clean check before publishing:
```
powershell -ExecutionPolicy Bypass -File scripts\check-clean-state.ps1 -VerboseOutput
```

## MCP Config (Claude Desktop)
Add to `mcpServers`:
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

## Project Layout
- `backend/`: AST/index/call chain/taint/scanner (gRPC+HTTP)
- `mcp/`: MCP server & tools
- `frontend/`: React frontend (optional)
- `proto/`: proto definitions & language-generated code
- `rules/`: common rules
- `release/`: distribution

## FAQ
- gRPC fails: ensure backend listening on `localhost:50051`
- HTTP gateway unavailable: ensure `-http-port` exposes `localhost:8080`
- Frontend port conflict: `$Env:PORT=3001; npm start`
- AI usage: server returns prompts; host LLM produces results; no API keys in server process

## License
MIT License