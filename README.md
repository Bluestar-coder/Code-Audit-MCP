# Code Audit MCP

[English](README.en.md) | [简体中文](README.md)

基于 Python + Go 的 AI 代码安全审计系统，支持 MCP（Model Context Protocol）接入。提供索引构建、调用链分析、污点追踪、规则驱动扫描、AI 解释与 PoC 提示，以及 OSV 漏洞检索。

## 特性
- 后端（Go，gRPC+HTTP）：AST/索引/调用链/污点/规则扫描
- MCP（Python）：统一工具接口，适配 MCP Host（如 Claude Desktop）
- 前端（React，可选）：仪表盘与漏洞列表
- 规则库（YAML）：内置常见漏洞规则

## 架构
- `backend/`：核心分析与服务接口（`cmd/server`、`internal`、`proto`）
- `mcp/`：MCP 服务器与工具入口（`python -m code_audit_mcp.server`）
- `frontend/`：React 前端（生产构建复制到 `release/frontend`）
- `proto/`：protobuf 定义与生成代码（Go/Python）
- `rules/`：YAML 规则库
- `scripts/`：`check-clean-state.ps1`、`build-release.ps1`
- `release/`：对外发布目录（`server.exe`、`frontend/`、`rules/common/`）

## 快速开始（Windows）
- 前置：安装 `Python 3.11+`、`Go 1.24+`（前端需 `Node.js 18+`）
- Python 环境（推荐虚拟环境）
```
cd mcp
python -m venv .venv
.\.venv\Scripts\Activate.ps1
pip install -e .
```
- 启动后端（默认 gRPC `localhost:50051`，HTTP `localhost:8080`）
```
cd backend
go mod download
go run ./cmd/server
# 或构建后运行：
# go build -o server.exe ./cmd/server
# .\server.exe -http-port 8080 -port 50051
```
- 启动前端（可选）
```
cd frontend
npm install
npm start   # 访问 http://localhost:3000
```

## 发布构建
- 一键脚本：
```
powershell -ExecutionPolicy Bypass -File scripts\build-release.ps1
```
- 输出内容：`release/server.exe`、`release/frontend/`、`release/rules/common/`
- 发布包说明见：`release/README-release.md`
- 发布前整洁检查：
```
powershell -ExecutionPolicy Bypass -File scripts\check-clean-state.ps1 -VerboseOutput
```

## MCP 配置（Claude Desktop）
将以下片段添加到 `mcpServers`：
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

## 目录结构
- `backend/`：AST/索引/调用链/污点/扫描（gRPC+HTTP）
- `mcp/`：MCP 服务器与工具
- `frontend/`：React 前端（可选）
- `proto/`：proto 定义与语言生成物
- `rules/`：通用漏洞规则
- `release/`：对外发布目录

## 常见问题
- gRPC 连接失败：确认后端运行并监听 `localhost:50051`
- HTTP 网关不可用：确认以 `-http-port` 启动并开放 `localhost:8080`
- Web UI 端口冲突：使用 `$Env:PORT=3001; npm start`
- AI 使用：服务器返回提示词，宿主 LLM 生成结果；无需在服务器进程配置模型密钥

## 许可
MIT License

