# ENG-N-004: Progressive delivery at scale

**Kind:** nice | **Domain:** eng | **Stack:** go+k8s

## Evidence from posting
Experience designing and operating progressive delivery systems at scale

## Rationale
Canary/blue-green maturity beyond basic rollout: promote a workload across
**multiple environments**, advancing only when automated criteria pass.

## Scope (Boundary Matrix D-03)

- **Owns:** multi-environment progressive delivery plans + automated promotion
  criteria (`dev` → `staging` → `prod`, gated per environment).
- **Does NOT own:** single-canary weight-step internals (that is ENG-E-024),
  CI DAGs, or reconcile finalizers. A weight field may be referenced but the
  E-024 weight-step machine is never the proof surface here.
- Criteria are **evaluated server-side** from observed metrics; a client cannot
  hand the service a pre-baked pass verdict.

## Acceptance demo
Multi-environment progressive delivery with automated promotion criteria.
`/v1/demo` blocks promotion under failing criteria, then auto-promotes to the
next environment once criteria pass.

## Run

```bash
make test         # go test -race ./...
make demo-local   # builds, runs on :18404, curls /v1/demo → demo-output.json
make demo         # docker compose variant
make down
```

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-n-004:local`, apply with kubectl/Kind).

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info`
- `POST /v1/plans` — create plan (`envs` ≥2, `criteria`)
- `POST /v1/plans/{id}/evaluate` — score observed metrics vs criteria
- `POST /v1/plans/{id}/promote` — advance to next env only when criteria pass
- `GET /v1/plans/{id}` — per-env status
- `GET|POST /v1/demo` — live proof: `environments`, `criteria_passed`, `auto_promoted`
- `GET /metrics`
