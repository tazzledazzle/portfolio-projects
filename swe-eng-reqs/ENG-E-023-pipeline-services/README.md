# ENG-E-023: Pipeline services

**Kind:** explicit | **Domain:** eng | **Stack:** go+compose

## Evidence from posting
Build pipeline services

## Rationale
Explicit deliverable distinct from owning CI conceptually.

## Acceptance demo
Pipeline service API: submit DAG, poll stages, stream logs.

## Run

```bash
make test
make demo
make down
```

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-e-023:local`, apply with kubectl/Kind).

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info`
- `GET|POST /v1/demo`
- `GET /metrics`
