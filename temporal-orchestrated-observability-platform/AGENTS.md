# Agent Instructions

## Project

**Temporal-Orchestrated Observability Platform** — AI workflows on Temporal, Kotlin + OpenTelemetry, LGTM/Grafana for operators.

## Before coding

1. Read `.planning/STATE.md` for current phase.
2. Read phase `CONTEXT.md` and execute `PLAN.md` only for active phase.
3. Read `docs/ARCHITECTURE.md` and relevant `docs/adr/*.md`.
4. Follow `.planning/research/PITFALLS.md` (no `workflow_id` metric labels, no non-deterministic workflow code).

## Conventions

- **Kotlin** JVM 21, Gradle Kotlin DSL
- **Temporal:** workflows in `:workflows`, worker in `:worker`, CLI in `:starter`
- **Telemetry:** `worker/.../telemetry/` — OpenTelemetry SDK
- **Infra:** `deploy/docker-compose.yml` — do not duplicate service definitions
- **Commits:** atomic per task; message references requirement ID (e.g. `FOUND-01: add compose stack`)

## Verification

- Run `./gradlew test` before claiming phase work complete
- Run `./scripts/smoke.sh` when touching Compose
- Run `./scripts/otel-smoke.sh` when worker telemetry changes (expects `:9464/metrics`)
- Run `./scripts/trace-smoke.sh` when collector/Jaeger/Prometheus config changes
- Run `./scripts/lgtm-sample-load.sh` when Grafana/Loki/Tempo/Promtail config changes
- Run `promtool check rules deploy/prometheus/alerts.yml` (via Docker) when alert rules change
- Create `{phase}-VERIFICATION.md` at phase end

## Telemetry (Phase 4+)

- Init: `OpenTelemetryConfig.init()` before `WorkerFactory.start()`
- Interceptors: `WorkerFactoryOptions.setWorkerInterceptors(TemporalTracingInterceptor(otel))`
- Env: `OTEL_EXPORTER_OTLP_ENDPOINT`, `OTEL_SERVICE_NAME`, `METRICS_PORT` (default 9464)
- Tests: `OTEL_TRACES_EXPORTER=none` via `OpenTelemetryConfig.initForTest()` (no Prometheus bind)

## GSD commands

| Command | When |
|---------|------|
| `/gsd-discuss-phase N` | Refine CONTEXT before execution |
| `/gsd-execute-phase N` | Run PLAN tasks |
| `/gsd-transition N` | Close phase, update PROJECT.md |

## Do not

- Add live LLM keys to repo
- Put business I/O inside workflow implementations
- Skip ADR process for stack changes
