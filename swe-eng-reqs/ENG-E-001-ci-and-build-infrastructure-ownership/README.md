# ENG-E-001: CI and build infrastructure ownership

**Kind:** explicit | **Domain:** eng | **Stack:** go+compose+k8s

## Evidence from posting
DevEx owns CI (continuous integration) and build infrastructure

## Rationale
Role is scoped to CI/build platforms, not general app features.

## Acceptance demo
Runnable multi-stage CI orchestrator with retries, events, and metrics.

## Run

```bash
make test
make demo
make down
```

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-e-001:local`, apply with kubectl/Kind).

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info`
- `GET|POST /v1/demo`
- `GET /metrics`
