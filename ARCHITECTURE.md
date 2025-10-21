# CodeAuditMcp 项目架构（重构版）

本架构旨在清晰分层与发布友好，降低模块耦合并统一命名。

## 模块总览
- `backend/`：Go 服务端（gRPC/HTTP）。典型结构：
  - `cmd/server/` 主入口
  - `internal/` 业务逻辑与服务实现（可选）
  - `proto/` Go 生成代码（pb.go），与 .proto 保持同步
  - `data/` 持久化数据（开发态，发布时不打包）
- `frontend/`：React 前端（生产构建输出到 `release/frontend`）。
- `mcp/`：Python MCP/插件侧，封装扫描任务、与后端/前端的交互。
- `proto/`：共享 `.proto` 定义与 Python 生成代码（`*_pb2.py`）。
- `rules/`：YAML 规则库（发布时复制到 `release/rules/common`）。
- `scripts/`：实用脚本（CI 检查、构建发布等）：
  - `check-clean-state.ps1` 发布前整洁检查
  - `build-release.ps1` 统一发布构建脚本
- `release/`：可对外发布的打包目录（二进制、前端静态、规则）。
- `.github/workflows/`：CI 工作流（自动运行整洁检查等）。

## 依赖与边界
- `backend` 与 `mcp` 均依赖 `proto` 的消息协议定义。
- `frontend` 只依赖后端公开 API（HTTP/REST 或 gRPC Web/网关）。
- `rules` 为纯数据，不与代码直接耦合；由后端/插件按需加载。

## 构建与发布
- 后端：在 `backend/` 运行 `go build`（脚本中自动选择入口）。
- 前端：在 `frontend/` 运行 `npm ci && npm run build`，输出复制到 `release/frontend`。
- 规则：从 `rules/common` 同步至 `release/rules/common`。
- 统一脚本：运行 `scripts/build-release.ps1` 完成上述全部流程。

## 命名与约定
- 统一采用模块名：`backend`、`frontend`、`mcp`，替代历史命名 `go-backend`、`web-ui`、`python-mcp`。
- 生成物与缓存不进版本库：
  - 前端构建：`frontend/build`（仅复制到 `release/frontend`）
  - Go 二进制：`backend/server.exe`（发布时位于 `release/`）
  - Python/Node/Go 缓存：由 `.gitignore` 与 `check-clean-state.ps1` 统一约束

## 路线图（可选优化）
- `.proto` 统一放置到 `proto/`，各语言生成物分别置于 `backend/proto/` 与 `mcp/proto/`，由脚本驱动生成。
- 后端托管前端静态资源，减少部署组件数。
- 增加版本化发布（如 `release/CodeAuditMcp-vX.Y.Z.zip`）。