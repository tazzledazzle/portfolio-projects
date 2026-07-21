# ENG-I-004: Multi-tenant consumer isolation

**Kind:** implicit | **Domain:** eng | **Stack:** go+compose

## Evidence from posting
Multi-tenant consumer isolation / noisy-neighbor protection

## Rationale
Platform consumers share capacity; quotas prevent cross-tenant steal and noisy neighbors.

## Acceptance demo
TenantScheduler enforces quotas and rate limits; quiet tenants keep scheduling when noisy tenants are limited.

## Run

```bash
make test
make demo-local
make down
```

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-i-004:local`).

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info`
- `PUT /v1/quotas` — set tenant quota (+ optional rate_limit)
- `POST /v1/schedule` — schedule units (requires tenant_id)
- `GET|POST /v1/demo` — live proof on `:18604`
- `GET /metrics`

## Boundary Matrix
Owns: tenant quotas + noisy-neighbor limits. Does **not** own multi-DC topology (E-010), chaos scenarios (H-001), or cost meter (H-002).
