# ENG-E-010: Multi-data-center operation (40+)

**Kind:** explicit | **Domain:** eng | **Stack:** go+compose

## Evidence from posting
across 40+ data centers with diverse infrastructure profiles

## Rationale
Geographic/ops complexity is explicit.

## Honesty
This slice is a **multi-DC simulator** (`multi_dc_simulator: true`). It does **not** operate real physical data centers.

## Acceptance demo
Simulated multi-DC control plane with ≥40 DCs, failure domains, and fan-out config (partial success for unhealthy sites).

## Run

```bash
make test
make demo-local
make down
```

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-e-010:local`, apply with kubectl/Kind).

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info` — simulator honesty labels
- `GET|POST /v1/dcs` — register/list data centers
- `POST /v1/fanout` — push config (partial OK)
- `GET /v1/domains` — failure domain grouping
- `GET|POST /v1/demo` — live proof on `:18510`
- `GET /metrics`

## Boundary Matrix
Owns: ≥40 DC topology + failure domains + fan-out. Does **not** own chaos blast (H-001), tenant quotas (I-004), or load sim (E-009).
