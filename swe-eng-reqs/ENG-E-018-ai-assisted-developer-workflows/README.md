# ENG-E-018: AI-assisted developer workflows

**Kind:** explicit | **Domain:** eng | **Stack:** python+compose

## Evidence from posting
AI-assisted developer workflows

## Rationale
End-to-end workflow design, not only model calls.

## Acceptance demo
Workflow: ingest failure → retrieve context → propose fix → require approval.

## Honesty labels

- `simulator=true`, `live_provider=false` — zero API keys; Python stdlib + pytest only
- Stages: `ingest` → `retrieve` → `propose` → `approval`
- Approval gate: status `awaiting_approval` until `approve()` → `approved`
- Does **not** own OfflineFixtureLLM product (E-015), ToolRegistry pedagogy (E-016), MCP protocol (E-017), or eval harness (H-005)

## Run

```bash
make test
env -u OPENAI_API_KEY -u ANTHROPIC_API_KEY make demo-local
```

The local proof listens on `127.0.0.1:18818` and writes `demo-output.json`.

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-e-018:local`, apply with kubectl/Kind).

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info`
- `GET|POST /v1/demo`
- `POST /v1/workflows`
- `POST /v1/workflows/{id}/advance`
- `POST /v1/workflows/{id}/approve`
- `GET /metrics`
