# Code Audit MCP

基于 Python+Go 的 AI 代码安全审计系统，支持 MCP（Model Context Protocol）接入，提供代码扫描、调用图分析、污点追踪、AI 解释与 PoC 生成、以及 OSV 漏洞查询。

## 快速开始（Windows）

- 前置：安装 `Python 3.12+`、`Go 1.21+`
- 启动 Go 后端（gRPC，默认 `localhost:50051`）
```
cd go-backend
go run cmd/server/main.go cmd/server/http_server.go
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
  - 功能：AI 解释代码（未配置时启发式回退）
- `generate_poc`
  - 入参：`vulnerability_id`，`language`，`context?`
  - 功能：AI 生成最小可复现 PoC（未配置时返回占位模板）
- `search_vulnerabilities`
  - 入参：`package_name`，`version?`，`ecosystem?`（如 `Go`/`PyPI`/`npm`）
  - 功能：调用 OSV API 返回已知漏洞，支持可选语义匹配

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
- AI 未启用：未设置 `ANTHROPIC_API_KEY` 时将启发式回退，不影响基本功能

## 许可
本项目采用 MIT 许可证。

