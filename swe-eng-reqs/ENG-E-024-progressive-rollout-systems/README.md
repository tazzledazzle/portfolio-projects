# ENG-E-024: Progressive rollout systems

**Kind:** explicit | **Domain:** eng | **Stack:** go+k8s

## Evidence from posting
progressive rollout systems

## Rationale
Explicit system to design and build in What You'll Do.

## Acceptance demo
Canary controller with weight steps, abort, and promote APIs.

## Run

```bash
make test
make demo-local
make demo   # requires Docker
make down
```

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-e-024:local`, apply with kubectl/Kind).

## Canary weights (default)
`0 → 10 → 50 → 100`. Abort forces weight 0; promote jumps to 100. Both are terminal (further Step rejected).

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info`
- `POST /v1/canaries` — start canary `{service, steps?}`
- `POST /v1/canaries/{id}/step` — advance to next weight
- `POST /v1/canaries/{id}/abort` — terminal abort (weight 0)
- `POST /v1/canaries/{id}/promote` — terminal promote (weight 100)
- `GET /v1/canaries/{id}` — status
- `GET|POST /v1/demo` — live proof (`canary_weights`, `abort_supported`, `promoted`/`aborted`)
- `GET /metrics`

## Demo-local
Port **18424** (`ADDR=:18424`). Does not implement multi-env PD plans (ENG-N-004 ownership).
