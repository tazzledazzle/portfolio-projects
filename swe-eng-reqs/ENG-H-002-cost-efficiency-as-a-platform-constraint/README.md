# ENG-H-002: Cost/efficiency as a platform constraint

**Kind:** hidden | **Domain:** eng | **Stack:** go+compose

## Evidence from posting
cloud-native scale + CI/build/artifacts at AI-cloud volumes + trade-offs

## Rationale
At this scale, uncontrolled CI/storage cost is existential.

## Acceptance demo
CostMeter records builds with duration/CPU/memory + cache_hit flag; server-computed `{cost_per_build_usd, cache_savings_pct}` on `:18702`. Does not implement CAS (N-010).

## Run

```bash
make test
make demo-local
make down
```

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-h-002:local`).

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info`
- `POST /v1/builds` — record build (duration/resources/cache_hit only)
- `GET /v1/cost-report` — server-computed cost + savings
- `GET|POST /v1/demo` — live proof on `:18702`
- `GET /metrics`

## Boundary Matrix
Owns: cost-per-build meter + cache-hit savings report. Does **not** own CAS cache impl (N-010) or tenant quotas (I-004). Cache hits are meter inputs only.
