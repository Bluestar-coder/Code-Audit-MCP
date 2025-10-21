# Security Policy

We take security seriously. Please follow the guidelines below when reporting vulnerabilities.

## Reporting a Vulnerability
- Do NOT open public issues for security vulnerabilities.
- Use one of the following confidential channels:
  - GitHub Security Advisories (preferred)
  - Email the maintainers: [REPLACE WITH CONTACT EMAIL]
- Include sufficient detail:
  - Affected components (e.g., `backend`, `mcp`, `frontend`)
  - Version/commit where the issue was found
  - Steps to reproduce and PoC (if available)
  - Impact assessment and suggested severity

## Disclosure Timeline
- We acknowledge reports within 3 business days.
- We coordinate validation, remediation, and fix preparation with reporters.
- After a fix is available, we will publish a new release and a public advisory.

## Supported Versions
- Active development project: we generally patch the latest release.
- Backports may be considered upon request and feasibility.

## Scope
- Analysis logic: AST / index / call chain / taint / rule scanner.
- Service interfaces: HTTP/gRPC endpoints, authentication/authorization (if applicable).
- MCP integration: tools interface and host interactions.
- Frontend: input handling and output rendering (e.g., XSS).
- Rules library: rule parsing/evaluation leading to security impact.

## Testing Guidance
- Use local or isolated environments when validating vulnerabilities.
- Avoid disruptive testing against production deployments.
- Provide minimal reproduction artifacts to help triage quickly.

Thank you for responsibly disclosing and helping improve the project’s security.