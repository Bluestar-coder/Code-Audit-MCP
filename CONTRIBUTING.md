# Contributing Guide

Thanks for considering contributing to Code Audit MCP! This guide outlines how to set up your environment, propose changes, and submit pull requests.

## Development Environment
- Windows recommended; Linux/macOS should also work.
- Install: `Python 3.11+`, `Go 1.24+`, optional `Node.js 18+` for the Web UI.
- Python virtualenv:
```
cd python-mcp
python -m venv .venv
.\.venv\Scripts\Activate.ps1
pip install -e .
pip install -e .[dev]   # optional
```
- Go backend:
```
cd go-backend
go mod download
go run ./cmd/server
# or build then run
# go build -o server.exe ./cmd/server
# .\server.exe -http-port 8080 -port 50051
```
- Web UI (optional):
```
cd web-ui
npm install
npm start
```

## Project Structure
- `go-backend/` — gRPC + HTTP services for AST/index/callchain/taint/scanner
- `python-mcp/` — MCP server and tools
- `web-ui/` — React + TypeScript frontend
- `proto/` — protobuf definitions + generated code
- `rules/` — scanning rules
- `docs/` — docs and assets

## Issue Reporting
- Search existing issues before opening a new one.
- Use clear titles and include minimal reproducible examples.
- For security issues, do NOT open public issues. See `SECURITY.md`.

## Branching & Commits
- Use feature branches: `feature/<short-topic>` or `fix/<short-topic>`.
- Follow Conventional Commits where possible:
  - `feat: add scan_vulnerabilities tool`
  - `fix: correct scan_code description`
  - `docs: add English README`
- Keep commits focused; avoid unrelated changes.

## Coding Style
- Go: `gofmt` and `golangci-lint` (if configured).
- Python: `black`, `ruff` (if configured).
- TypeScript: `prettier`, `eslint` (if configured).
- Add tests when applicable (`python-mcp/src/test_*.py`).

## Pull Request Checklist
- PR title is clear and follows Conventional Commits (preferred).
- Include a description of the change and motivation.
- Ensure Go backend and Python MCP server still run.
- Update docs (`README.md`, `README.en.md`) if behavior or endpoints changed.
- Add or update tests when relevant.

## Reviews & Merging
- PRs require at least one approval.
- Squash merge is recommended to keep a clean history.
- Maintainers may request changes to align with project scope and quality.

## Releases
- Follow semantic versioning with tags `vX.Y.Z` when publishing.
- Maintain a `CHANGELOG.md` if the project grows; optional initially.

## Questions
Have questions? Open a discussion or issue with the label `question`.