# ENG-N-010: Artifact / build caching

**Kind:** nice | **Domain:** eng | **Stack:** go+compose

## Evidence from posting
caching

## Rationale
Performance layer distinct from storage.

## Acceptance demo
CAS cache with namespace isolation and LRU eviction.

## Run

```bash
make test
make demo
make down
```

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-n-010:local`, apply with kubectl/Kind).

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info`
- `GET|POST /v1/demo`
- `GET /metrics`
