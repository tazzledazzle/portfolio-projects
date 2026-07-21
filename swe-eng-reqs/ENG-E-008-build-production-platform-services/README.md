# ENG-E-008: Build production platform services

**Kind:** explicit | **Domain:** eng | **Stack:** go+compose+k8s

## Evidence from posting
Design and build the core services that power developer productivity

## Rationale
Hands-on IC contribution required alongside leadership. This slice owns **production runtime quality** (request IDs, structured errors, `/metrics`) — not HPA/manifest depth (ENG-E-013), language OR meta (ENG-E-012), or IDP catalog (ENG-E-005).

## Acceptance demo
Live `/v1/demo` on port **18508** proves `metrics_exposed`, `request_id`, and `structured_error` (no stack traces / secrets).

## Run

```bash
make test
make demo-local
make demo   # optional Docker Compose
make down
```

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-e-008:local`, apply with kubectl/Kind).

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info`
- `GET|POST /v1/echo` — echoes payload; returns `X-Request-ID` + body `request_id`
- `GET|POST /v1/demo` — live production-quality proof
- `GET /metrics` — Prometheus-style counters
