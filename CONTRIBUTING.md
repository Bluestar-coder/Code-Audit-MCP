# Contributing Guide

Thanks for contributing to Code Audit MCP! This guide covers environment setup, proposing changes, and submitting PRs.

## Development Environment
- Windows recommended; Linux/macOS should also work.
- Install: `Python 3.11+`, `Go 1.24+`, optional `Node.js 18+` for the frontend.
- Python virtualenv:
```
cd mcp
python -m venv .venv
.\.venv\Scripts\Activate.ps1
pip install -e .
# optional dev deps
pip install -e .[dev]
```
- Go backend:
```
cd backend
go mod download
go run ./cmd/server
# or build then run
# go build -o server.exe ./cmd/server
# .\server.exe -http-port 8080 -port 50051
```
- Frontend (optional):
```
cd frontend
npm install
npm start
```

## Project Structure
- `backend/` — gRPC + HTTP services for AST/index/callchain/taint/scanner
- `mcp/` — MCP server and tools
- `frontend/` — React app
- `proto/` — protobuf definitions + generated code
- `rules/` — scanning rules
- `scripts/` — build & cleanliness scripts

## Issue Reporting
- Search existing issues before opening new ones.
- Use clear titles and minimal reproducible examples.
- For security issues, do NOT open public issues. See `SECURITY.md`.

## Branching & Commits
- Use feature branches: `feature/<topic>` or `fix/<topic>`.
- Conventional Commits (preferred):
  - `feat: add scan_vulnerabilities tool`
  - `fix: correct scan_code description`
  - `docs: update README for new structure`
- Keep commits focused; avoid unrelated changes.

## Coding Style
- Go: `gofmt` and `golangci-lint` (if configured).
- Python: `black`, `ruff` (if configured).
- TypeScript: `prettier`, `eslint` (if configured).
- Add tests when applicable.

## Pull Request Checklist
- Title clear and (preferably) Conventional Commits.
- Describe change and motivation.
- Ensure backend and MCP server still run.
- Update docs (`README.md`, `README.en.md`, `ARCHITECTURE.md`) if behavior or endpoints changed.

## Reviews & Merging
- At least one approval required.
- Squash merge recommended.
- Maintainers may request changes for scope/quality alignment.

## Releases
- Use semantic versioning tags `vX.Y.Z` when publishing.
- Consider maintaining `CHANGELOG.md` as the project grows (optional initially).

## Questions
Have questions? Open a discussion or issue with the `question` label.