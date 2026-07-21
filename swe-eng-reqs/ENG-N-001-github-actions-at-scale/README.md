# ENG-N-001: GitHub Actions at scale

**Kind:** nice | **Domain:** eng | **Stack:** go+compose

## Important: This is a SIMULATOR

This service simulates GitHub Actions runner behavior for portfolio demonstration.
It does NOT connect to GitHub, does NOT run actual workflows, and is NOT production-ready.

Purpose: Demonstrate understanding of Actions architecture (matrix, runners, secrets).

## Evidence from posting
GitHub Actions, Buildkite, Bazel, or similar build systems

## Rationale
Common CI product signal.

## Acceptance demo
Reusable Actions-compatible workflow runner simulator with matrix + secrets model.

## Run

```bash
make test
make demo-local
make demo   # requires Docker
make down
```

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-n-001:local`, apply with kubectl/Kind).

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info`
- `GET|POST /v1/demo`
- `POST /v1/workflows` — expand matrix and enqueue jobs
- `POST /v1/runners` — register runner
- `GET /v1/runners/:id/claim` — claim next job (204 if empty)
- `POST /v1/jobs/:id/complete` — report job completion
- `GET /metrics`
