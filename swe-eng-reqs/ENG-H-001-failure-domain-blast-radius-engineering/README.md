# ENG-H-001: Failure-domain / blast-radius engineering

**Kind:** hidden | **Domain:** eng | **Stack:** go+compose

## Evidence from posting
Failure-domain / blast-radius engineering for multi-tenant platforms

## Rationale
Chaos in one domain must not cascade; prove containment with unaffected tenants/domains.

## Acceptance demo
BlastEngine runs a chaos scenario in one failure domain; server-computed blast radius lists affected vs unaffected sets with `contained=true`.

## Run

```bash
make test
make demo-local
make down
```

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-h-001:local`).

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info`
- `POST /v1/chaos` — inject scenario into one domain
- `GET /v1/blast-radius` — server-computed affected/unaffected sets
- `GET|POST /v1/demo` — live proof on `:18701`
- `GET /metrics`

## Boundary Matrix
Owns: chaos containment proof. Does **not** own multi-DC fan-out (E-010) or quota math (I-004). Uses local in-memory domains/tenants only.
