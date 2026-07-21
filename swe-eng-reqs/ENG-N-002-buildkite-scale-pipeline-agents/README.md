# ENG-N-002: Buildkite-scale pipeline agents

**Kind:** nice | **Domain:** eng | **Stack:** go+compose

## Important: This is a SIMULATOR

This service simulates Buildkite agent/queue behavior for portfolio demonstration.
It does NOT connect to Buildkite, does NOT run actual pipelines, and is NOT production-ready.

Purpose: Demonstrate understanding of Buildkite architecture (agents, dynamic pipelines, concurrency groups).

## Evidence from posting
Buildkite

## Rationale
Breadth across CI vendors valued.

## Acceptance demo
Agent/queue model with dynamic pipelines and concurrency limits.

## Run

```bash
make test
make demo-local
make demo   # requires Docker
make down
```

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-n-002:local`, apply with kubectl/Kind).

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info`
- `GET|POST /v1/demo`
- `POST /v1/agents` — register agent
- `GET /v1/agents/:id/poll` — poll/claim next job (204 if empty/limited)
- `POST /v1/pipelines/:id/upload` — dynamic pipeline YAML
- `POST /v1/jobs/:id/complete` — complete job and release concurrency
- `GET /v1/concurrency` — concurrency group statuses
- `GET /metrics`
