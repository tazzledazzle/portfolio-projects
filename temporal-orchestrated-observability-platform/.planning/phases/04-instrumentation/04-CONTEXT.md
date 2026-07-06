# Phase 4: Instrumentation (Kotlin + OTel) — Context

**Gathered:** 2026-05-19  
**Planned:** 2026-05-20

<domain>

## Phase boundary

Deliver **OTEL-01** through **OTEL-04** on the Kotlin worker:

- Global `OpenTelemetry` SDK with resource attributes (`service.name=ai-temporal-worker`).
- Custom Temporal `WorkerInterceptor` creating activity spans with Temporal context (ADR-004).
- Prometheus exporter on `:9464/metrics`.
- JSON logging with trace correlation.
- OkHttp OTel instrumentation for LLM stub calls.

**Out of scope:** Collector → Jaeger wiring (Phase 5); Grafana dashboards (Phase 6).

Phase 4 exports OTLP to the Phase 1 collector (`deploy/otel-collector/config.yaml`). Jaeger UI proof waits for Phase 5.

</domain>

<decisions>

- **D-01:** `service.name=ai-temporal-worker`, `service.version` from Gradle `0.1.0-SNAPSHOT`.
- **D-02:** Metric names: `workflow.completed`, `activity.duration`, `llm.request.duration` (seconds histogram).
- **D-03:** Metric labels: `workflow_type`, `activity_type`, `status` — never `workflow_id` or `run_id`.
- **D-04:** Logback JSON encoder with MDC keys `trace_id`, `span_id`, `workflow_id`, `run_id`, `workflow_type`, `activity_type`.
- **D-05:** **Custom** `WorkerInterceptor` — `io.temporal:temporal-opentelemetry` is not on Maven Central for SDK 1.25.2 (see `04-RESEARCH.md`).
- **D-06:** OTel BOM `1.43.0`; Prometheus scrape port `METRICS_PORT` default `9464`.
- **D-07:** Span naming: `activity.<activityType>`, `llm.complete` — no high-cardinality URLs.
- **D-08:** Two execution waves: **04-01** traces/metrics, **04-02** logs/HTTP/smoke.

</decisions>

<execution_plans>

| Plan | Wave | Requirements |
|------|------|----------------|
| [04-01-PLAN.md](04-01-PLAN.md) | 1 | OTEL-01, OTEL-02 |
| [04-02-PLAN.md](04-02-PLAN.md) | 2 | OTEL-03, OTEL-04 |

</execution_plans>

<canonical_refs>

- `docs/adr/0002-kotlin-worker-opentelemetry.md`
- `docs/adr/0004-workflow-trace-correlation.md`
- `.planning/phases/04-instrumentation/04-RESEARCH.md`
- `.planning/research/PITFALLS.md`

</canonical_refs>
