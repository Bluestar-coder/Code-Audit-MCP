# Security Policy

We take security seriously. Please follow the guidelines below when reporting vulnerabilities.

## Reporting a Vulnerability
- Do NOT open public issues for security vulnerabilities.
- Instead, use one of the following confidential channels:
  - GitHub Security Advisories (preferred)
  - Email the maintainers: [REPLACE WITH CONTACT EMAIL]
- Provide as much detail as possible:
  - Affected components (e.g., `go-backend`, `python-mcp`, `web-ui`)
  - Version/commit where the issue was found
  - Steps to reproduce, PoC (if available)
  - Impact assessment and suggested severity

## Disclosure Timeline
- We aim to acknowledge reports within 3 business days.
- We will work with you to validate the issue and prepare a fix.
- Once a fix is available, we will publish a new release and a public advisory.

## Supported Versions
- This is an active development project. We generally patch the latest release.
- If you require backports, please communicate your needs in the initial report.

## Scope
- Vulnerabilities in analysis logic (AST/index/call chain/taint/scanner)
- HTTP/gRPC endpoints and authentication/authorization (if applicable)
- Web UI input handling and output rendering (e.g., XSS)

Thanks for responsibly disclosing and helping improve the project’s security.