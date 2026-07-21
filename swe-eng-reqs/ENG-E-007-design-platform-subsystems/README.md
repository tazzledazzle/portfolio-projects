# ENG-E-007: Design platform subsystems

**Kind:** explicit | **Domain:** eng | **Stack:** go+compose

## Evidence from posting
Design and build the core services... Drive architectural decisions

## Rationale
Expected to own design, not only implement tickets. This slice owns an ADR with
≥2 alternatives plus a thin runnable skeleton. It does **not** own full
production metrics (ENG-E-008), HPA packaging (ENG-E-013), or Phase 7 mentoring
kits.

## Acceptance demo
ADR-backed service skeleton with documented alternatives and trade-offs. Live
`/v1/demo` on port **18507** proves `adr_count`, `alternatives≥2`, and
`decision_recorded`.

## ADR

See `adr/0001-platform-subsystem.md` (original paraphrase; no copyrighted book
text). Endpoints: `GET /v1/adrs`, `GET /v1/skeleton`.

## Run

```bash
make test
make demo-local
make demo
make down
```

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-e-007:local`, apply with kubectl/Kind).

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info`
- `GET /v1/adrs`
- `GET /v1/skeleton`
- `GET|POST /v1/demo`
- `GET /metrics`
