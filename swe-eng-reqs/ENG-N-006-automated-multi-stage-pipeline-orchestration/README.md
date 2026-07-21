# ENG-N-006: Automated multi-stage pipeline orchestration

**Kind:** nice | **Domain:** eng | **Stack:** go+compose

## Evidence from posting
automated multi-stage pipeline orchestration

## Rationale
Structured promote-through-environments automation: a release advances through
environment stages (`dev` → `staging` → `prod`) **only when a quality gate
allows it**.

## Scope (Boundary Matrix D-03 / D-09)

This service is **release/environment stage orchestration**, not a CI DAG.

- **Owns:** advancing release/env stages only when `gate == allow`.
- **Does NOT own:** Phase 2 CI stages (`lint` → `unit` → `build` → `publish`),
  canary weight math, or live PromQL. Those belong to ENG-E-001/ENG-E-023
  (CI) and ENG-E-024/ENG-N-005 (canary/gate).
- **Gate coupling:** the `GateEvaluator` is an **embedded stub** — this folder
  runs independently and never HTTP-calls ENG-N-005 (D-05, D-11).

## Acceptance demo
Orchestrator that advances stages only when gates pass. `/v1/demo` denies once
(proves `blocked_on_deny`) then allows to terminal (proves `stages_advanced`),
always consulting the gate (`gate_required`).

## Run

```bash
make test         # go test -race ./...
make demo-local   # builds, runs on :18406, curls /v1/demo → demo-output.json
make demo         # docker compose variant
make down
```

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-n-006:local`, apply with kubectl/Kind).

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info`
- `POST /v1/orchestrations` — create (optional `stages`, `slo_id`)
- `POST /v1/orchestrations/{id}/tick` — evaluate gate, advance one stage on allow
- `GET /v1/orchestrations/{id}` — current stage/state
- `GET|POST /v1/demo` — live proof: `stages_advanced`, `blocked_on_deny`, `gate_required`
- `GET /metrics`
