# ENG-E-006: Agentic AI tooling for DevEx

**Kind:** explicit | **Domain:** eng | **Stack:** python+compose

## Evidence from posting
agentic AI tooling that is redefining how engineers move from idea to production

## Rationale
AI is in the team charter, not optional curiosity.

## Acceptance demo
Diagnose pipeline fixtures and propose safe actions with `executed=false`.

## Honesty labels

- `offline_fixture_llm=true` / `simulator=true`
- `live_provider=false` — zero API keys; Python stdlib + pytest only
- `execute_mutating=false` — propose-only; mutating actions set `requires_approval=true`
- Does **not** own agent-framework pedagogy (E-016), MCP auth (E-017), full policy gateway (N-015), or ROI math (N-014)

## Run

```bash
make test
env -u OPENAI_API_KEY -u ANTHROPIC_API_KEY make demo-local
```

The local proof listens on `127.0.0.1:18806` and writes `demo-output.json`.

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-e-006:local`, apply with kubectl/Kind).

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info`
- `GET|POST /v1/demo`
- `POST /v1/diagnose`
- `GET /metrics`
