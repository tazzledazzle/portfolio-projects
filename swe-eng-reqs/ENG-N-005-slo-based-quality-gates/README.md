# ENG-N-005: SLO-based quality gates

**Kind:** nice | **Domain:** eng | **Stack:** go+compose

## Evidence from posting
SLO-based quality gates

## Rationale
Release decisions tied to service objectives.

## Acceptance demo
Gate service deny/allow based on PromQL-inspired burn rate with evidence payload.

## SIMULATOR / PromQL-inspired

This slice is a **PromQL-inspired simulator**. It does **NOT** connect to Prometheus.
Burn rates are computed from in-memory series fixtures (`burn_rate = error_ratio / (1 - objective)`).
Deny when **both** short and long windows exceed the burn threshold (multi-window AND).

## Run

```bash
make test
make demo-local
make demo   # requires Docker
make down
```

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-n-005:local`, apply with kubectl/Kind).

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info` — includes `promql_inspired: true`, `simulator: true`
- `PUT /v1/slos/{id}` — `{objective, threshold}`
- `POST /v1/series/{id}` — ingest window samples
- `POST /v1/gates/evaluate` — `{slo_id}` → allow/deny + server-side evidence (client `burn_rate` ignored)
- `GET|POST /v1/demo` — live proof of allow **and** deny paths
- `GET /metrics`

## Demo-local
Port **18405** (`ADDR=:18405`).
