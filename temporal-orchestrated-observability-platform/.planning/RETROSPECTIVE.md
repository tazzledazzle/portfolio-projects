# Project Retrospective

Living document — append a section per shipped milestone.

## Milestone: v1.0 — Reference Platform

**Shipped:** 2026-05-20 (archived 2026-06-22)  
**Phases:** 7 | **Plans:** 8

### What Was Built

- Reproducible Compose stack with smoke validation (Foundation)
- Kotlin Temporal worker with PingWorkflow + three AI reference workflows
- OpenTelemetry instrumentation: interceptors, Prometheus metrics, JSON logs, OkHttp propagation
- Jaeger + Prometheus backends; OTel Collector CI validation
- LGTM integration: Loki, Tempo, Grafana dashboards with log↔trace correlation
- Operator runbooks and Prometheus alert rules with manual validation

### What Worked

- Bottom-up roadmap aligned with architecture diagram — each phase built on prior infra
- Verification docs (`*-VERIFICATION.md`) provided durable evidence without GSD SUMMARY artifacts
- ADR-driven decisions (especially workflow_id label cardinality) prevented observability pitfalls
- Smoke scripts (`smoke.sh`, `otel-smoke.sh`, `trace-smoke.sh`, `lgtm-sample-load.sh`) enabled repeatable validation

### What Was Inefficient

- GSD plan/summary tracking never backfilled — milestone tooling reported 0% until manual archive
- ROADMAP progress table drifted from actual completion state
- Single git commit history limits per-phase timeline granularity

### Patterns Established

- Temporal workflows in `:workflows`, worker in `:worker`, CLI in `:starter`
- OTel init before `WorkerFactory.start()`; interceptors via `WorkerFactoryOptions`
- No `workflow_id` on Prometheus metric labels (traces/logs only)
- All infra in `deploy/docker-compose.yml` — no duplicate service definitions

### Key Lessons

- Verification docs are sufficient for portfolio projects even when GSD SUMMARY.md is skipped
- Dual Jaeger + Tempo export supports both fast debug and Grafana correlation
- Operator docs (`OPERATIONS.md`) should ship with alert rules, not after

### Cost Observations

- Milestone executed primarily via quality model profile
- 7 phases over ~2 days (2026-05-19 → 2026-05-20)

## Cross-Milestone Trends

| Milestone | Phases | Plans | Requirements | Shipped |
|-----------|--------|-------|--------------|---------|
| v1.0 Reference Platform | 7 | 8 | 22/22 | 2026-05-20 |
