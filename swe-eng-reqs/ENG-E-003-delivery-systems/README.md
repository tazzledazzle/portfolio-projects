# ENG-E-003: Delivery systems

**Kind:** explicit | **Domain:** eng | **Stack:** go+compose+k8s

## Evidence from posting
delivery systems

## Rationale
Release/delivery automation is first-class DevEx ownership.

## Acceptance demo
Multistage promote pipeline with audit trail and rollback hooks.

## Run

```bash
make test
make demo-local
make demo   # requires Docker
make down
```

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-e-003:local`, apply with kubectl/Kind).

## Release stages
`dev` → `staging` → `prod` (exactly one step per promote). Not CI stages (`lint`/`unit`/`build`/`publish`).

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info`
- `POST /v1/releases` — create release `{app, version, stage}`
- `POST /v1/releases/{id}/promote` — advance one stage; appends audit
- `POST /v1/releases/{id}/rollback` — invoke registered rollback hooks once; appends audit
- `GET /v1/releases/{id}/audit` — append-only audit trail
- `GET|POST /v1/demo` — live proof (`promoted`, `audit_entries`, `rollback_invoked`)
- `GET /metrics`

## Demo-local
Port **18403** (`ADDR=:18403`).
