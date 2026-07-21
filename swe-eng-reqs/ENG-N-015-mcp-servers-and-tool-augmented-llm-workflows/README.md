# ENG-N-015: MCP servers and tool-augmented LLM workflows

**Kind:** nice | **Domain:** eng | **Stack:** python+compose

## Evidence from posting
including MCP servers, tool-augmented LLM workflows

## Rationale
Deeper than familiarity: implement MCP + tool-augmented flows.

## Acceptance demo
MCP server + policy gateway requiring human approval for mutating tools.

## Simulator honesty

This slice is `mcp_inspired=true`, `mcp_sdk=false`, and
`policy_gateway=true`. It implements a stdlib `tools/list` and `tools/call`
subset, not the official MCP SDK. Mutating calls are denied until a human
approval grant is issued for the exact tool name and arguments.

## Run

```bash
make test
make demo-local
```

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-n-015:local`, apply with kubectl/Kind).

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info`
- `GET|POST /v1/demo`
- `POST /v1/mcp`
- `POST /v1/approvals`
- `GET /metrics`
