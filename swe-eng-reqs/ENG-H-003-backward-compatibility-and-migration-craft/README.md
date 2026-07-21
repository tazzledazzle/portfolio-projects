# ENG-H-003: Backward compatibility and migration craft

**Kind:** hidden | **Domain:** eng | **Stack:** go+compose

## Evidence from posting
multi-team IDP/CI/artifacts; multi-quarter work

## Rationale
Platforms cannot break all consumers; migrations are the job.

## Acceptance demo
API v1→v2 dual-write under one mutex; v1 keeps `name`, v2 uses `display_name`; live proof `{dual_write, v1_readable, v2_readable, compat_pass}` on `:18703`.

## Run

```bash
make test
make demo-local
make down
```

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-h-003:local`).

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info`
- `POST /v1/migrate/dual-write` — enable dual-write
- `POST /v1/items` / `GET /v1/items/{id}` — legacy shape
- `POST /v2/items` / `GET /v2/items/{id}` — renamed field shape
- `GET|POST /v1/demo` — live proof on `:18703`
- `GET /metrics`

## Boundary Matrix
Owns: API v1→v2 dual-write + compatibility tests. Does **not** own OpenAPI/rate-limit craft (E-021) or IDP catalog (E-005).
