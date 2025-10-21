# Code Audit MCP

基于 Python+Go 的 AI 代码安全审计系统，支持 MCP（Model Context Protocol）接入，提供代码扫描、调用图分析、污点追踪、AI 解释与 PoC 生成、以及 OSV 漏洞查询。

## 快速开始（Windows）

- 前置：安装 `Python 3.11+`、`Go 1.24+`（以及可选 `Node.js 18+` 用于 Web UI）
- Python 环境（推荐虚拟环境）
```
cd python-mcp
python -m venv .venv
.\.venv\Scripts\Activate.ps1
pip install -e .            # 安装核心依赖
# 可选：开发依赖
pip install -e .[dev]
```
- 启动 Go 后端（gRPC，默认 `localhost:50051`）
```
cd go-backend
go mod download
go run cmd/server/main.go cmd/server/http_server.go
# 或构建后运行：
# go build -o server.exe ./cmd/server
# .\server.exe
```
- 启动 Web UI（可选）
```
cd web-ui
npm install
npm start   # 访问 http://localhost:3000
```
- 运行 Python 工具快速验证
```
cd python-mcp/src
python test_scan.py        # 索引与扫描示例
python test_call_graph.py  # 调用图构建示例
python test_taint.py       # 污点路径追踪示例
python test_poc.py         # PoC 生成示例（AI 可选）
python test_vuln.py        # OSV 漏洞查询示例
```

## MCP 服务器与工具
MCP 服务器入口：`python -m code_audit_mcp.server`（适用于 MCP Host，如 Claude Desktop）

- `scan_code`
  - 入参：`path`（文件或目录），`language?`
  - 功能：调用 Go Indexer 构建索引并统计函数/类/变量
- `analyze_call_graph`
  - 入参：`path`，`entry_point?`
  - 功能：调用 Go CallChainAnalyzer 构建调用图
- `trace_taint`
  - 入参：`source`，`sink`
  - 功能：调用 Go TaintAnalyzer 追踪从源到汇的数据流
- `explain_code`
  - 入参：`code`，`language?`
  - 功能：返回宿主 LLM 的“建议提示词”（不在服务器侧调用 SDK）
- `generate_poc`
  - 入参：`vulnerability_id`，`language`，`context?`
  - 功能：返回生成最小可复现 PoC 的“建议提示词”（不在服务器侧调用 SDK）
- `search_vulnerabilities`
  - 入参：`package_name`，`version?`，`ecosystem?`（如 `Go`/`PyPI`/`npm`）
  - 功能：调用 OSV API 返回已知漏洞，支持可选语义匹配
- `scan_vulnerabilities`
  - 入参：`file_path?`，`language?`，`content?`，`rule_ids?`
  - 功能：通过 HTTP `/api/scan` 调用 Go 漏洞检测服务，返回扫描 `findings` 与统计

## 可选 AI 与语义检索
- 设置环境变量以启用 AI
```
$Env:ANTHROPIC_API_KEY="your-key"
$Env:CLAUDE_MODEL="claude-3-5-sonnet-latest"   # 可选
```
- 语义检索：安装 `txtai` 后自动启用（未安装则关键词回退）

## 目录结构
- `go-backend/`：AST 解析、索引、调用链与污点分析（gRPC）
- `python-mcp/`：MCP 服务器与工具实现、AI/语义检索
- `proto/`：proto 定义及 Python 生成文件
- `web-ui/`：前端（可选）

## 常见问题
- gRPC 连接失败：确认 Go 后端已启动并监听 `localhost:50051`
- OSV 查询超时：检查网络或稍后重试；失败时会返回错误说明
- AI 调用：服务器不直接调用 LLM，工具返回提示词交由宿主生成结果；无需配置 `ANTHROPIC_API_KEY`。

## Claude Desktop 配置
将以下片段添加到 Claude Desktop 的配置（`mcpServers`）中：
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
说明：
- `command` 与 `args` 指向 MCP 服务器入口。
- gRPC 连接默认 `localhost:50051`，可在 Go 后端启动时自定义端口。
- AI 由宿主驱动：服务器返回提示词，宿主 LLM 生成结果；无需在服务器进程中配置模型密钥。

## 端到端示例
### 1) 代码扫描 / 调用图 / 污点
确保 Go 后端已启动后运行：
```
cd python-mcp/src
python test_scan.py
python test_call_graph.py
python test_taint.py
```
示例输出（截断）：
```
{"path":"E:\\Code\\CodeAuditMcp\\go-backend\\internal\\indexer\\service.go","functions":10,"classes":5,"variables":50}
{"nodes":25,"edges":40,"path":"...service.go"}
{"paths_found":1,"source":"user_input","sink":"os/exec"}
```

### 2) 漏洞检索 + PoC 生成
```
python test_vuln.py
python test_poc.py
```
说明：
- `test_vuln.py` 使用 OSV 查询，例如 `CVE-2023-29401`、`GO-2023-1737`。
- `test_poc.py` 将返回宿主 LLM 的“建议提示词”，请在宿主中使用该提示词生成 PoC。

## 许可
本项目采用 MIT 许可证。

