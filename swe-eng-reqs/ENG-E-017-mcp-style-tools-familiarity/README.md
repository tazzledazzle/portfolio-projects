# ENG-E-017: MCP-style tools familiarity

**Kind:** explicit | **Domain:** eng | **Stack:** python+compose

## Evidence from posting
MCP-style tools

## Rationale
Model Context Protocol style tools called out specifically.

## Acceptance demo
MCP server exposing read-only DevEx tools with auth and audit log.

## SIMULATOR / mcp_inspired

This slice is an in-process, `mcp_inspired` simulator of `tools/list` and
`tools/call`. It is deliberately read-only, uses token-inspired scopes, and
records append-only redacted audits. `mcp_sdk=false`: it does not use or claim
conformance with the official MCP SDK or transport protocol.

## Run

```bash
make test
make demo-local  # listens on :18817, no API keys or Docker required
```

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-e-017:local`, apply with kubectl/Kind).

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info`
- `GET|POST /v1/demo`
- `POST /v1/mcp` (`tools/list` or authenticated `tools/call`)
- `GET /metrics`
