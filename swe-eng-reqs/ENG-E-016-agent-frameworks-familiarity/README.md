# ENG-E-016: Agent frameworks familiarity

**Kind:** explicit | **Domain:** eng | **Stack:** python+compose

## Evidence from posting
such as agent frameworks, MCP-style tools, or AI-assisted developer workflows

## Rationale
Agent frameworks are an explicit example of expected AI tooling fluency.

## Acceptance demo
Minimal agent loop with tool registry, planning step, and deterministic eval.

## Agent framework honesty

- `agent_framework_inspired=true`
- `simulator=true`
- `live_provider=false`
- Implements Plan-Execute, act/observe tracing, and a typed `ToolRegistry` using Python stdlib.
- Does not embed LangChain, CrewAI, AutoGen, MCP, or a live LLM provider.
- Scope is framework pedagogy only; DevEx diagnosis, MCP protocol, policy gateways, and ROI belong to other requirement slices.

## Run

```bash
make test
make demo-local
```

The local proof listens on `127.0.0.1:18816` and writes `demo-output.json`.

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-e-016:local`, apply with kubectl/Kind).

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info`
- `GET|POST /v1/demo`
- `POST /v1/agent/run`
- `GET /metrics`
