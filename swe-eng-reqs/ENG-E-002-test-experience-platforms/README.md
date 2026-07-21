# ENG-E-002: Test experience platforms

**Kind:** explicit | **Domain:** eng | **Stack:** go+compose

## Evidence from posting
scope spans ... test experience

## Rationale
Test platforms are distinct from CI runners: flake visibility, quarantine, DX.

## Acceptance demo
Ingest JUnit results, compute flake scores, quarantine flaky tests from gates.

## Run

```bash
make test
make demo
make down
```

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-e-002:local`, apply with kubectl/Kind).

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info`
- `GET|POST /v1/demo`
- `GET /metrics`
