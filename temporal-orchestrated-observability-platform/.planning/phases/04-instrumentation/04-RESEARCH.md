# Phase 4 Research — Kotlin + OpenTelemetry

**Researched:** 2026-05-20
**Phase:** 04-instrumentation
**Requirements:** OTEL-01 … OTEL-04

## Summary

Instrument the **Kotlin worker** (`:worker`) as the single telemetry boundary: OTel SDK in-process, **custom Temporal interceptors** (no `io.temporal:temporal-opentelemetry` artifact on Maven Central for SDK 1.25.2), Prometheus scrape on **`:9464/metrics`**, JSON logs with MDC, and OkHttp child spans for the LLM stub.

Phase 5 wires collector → Jaeger/Tempo; Phase 4 only needs OTLP export to the existing Compose collector (`deploy/otel-collector/config.yaml`) plus local Prometheus scrape already defined in `deploy/prometheus/prometheus.yml`.

## Technology choices

| Area             | Choice                                                                                                     | Rationale                                                                     |
|------------------|------------------------------------------------------------------------------------------------------------|-------------------------------------------------------------------------------|
| OTel API/SDK     | BOM `1.43.0`                                                                                               | Aligns with `.planning/research/STACK.md`; stable OTLP + Prometheus exporters |
| Trace export     | OTLP gRPC → `OTEL_EXPORTER_OTLP_ENDPOINT`                                                                  | Collector from Phase 1 already listens on `4317`                              |
| Metrics export   | `PrometheusHttpServer` on `METRICS_PORT` (default `9464`)                                                  | Matches `prometheus.yml` `ai-worker` job                                      |
| Temporal tracing | Custom `WorkerInterceptor` + `ActivityInboundCallsInterceptorBase` + `WorkflowInboundCallsInterceptorBase` | Official `temporal-opentelemetry` module not published for 1.25.2             |
| HTTP client      | `opentelemetry-okhttp-3.0` instrumentation                                                                 | W3C `traceparent` on LLM stub calls (OTEL-04)                                 |
| Logging          | `logstash-logback-encoder` + manual MDC in interceptor                                                     | JSON one-liners; `trace_id` from active span context                          |

## Temporal interceptor pattern

```text
WorkerFactory.newWorker(queue, WorkerOptions)
  └── WorkerInterceptor
        ├── interceptWorkflow → WorkflowTracingInterceptor
        │     └── on success/failure: workflow.completed counter
        └── interceptActivity → ActivityTracingInterceptor
              ├── start span: name = activity_type
              ├── attributes: workflow_id, run_id, workflow_type, activity_type, task_queue
              ├── MDC: same keys + trace_id (hex, no dashes)
              ├── record activity.duration histogram
              └── end span (status from exception)
```

**Cardinality rule (ADR-004):** `workflow_id` / `run_id` on **spans and logs only** — never on Prometheus metric attributes.

## Metric instrument names

| Instrument             | Type                | Attributes                                 |
|------------------------|---------------------|--------------------------------------------|
| `activity.duration`    | Histogram (seconds) | `workflow_type`, `activity_type`, `status` |
| `workflow.completed`   | Counter             | `workflow_type`, `status`                  |
| `llm.request.duration` | Histogram (seconds) | `status` (optional `model=stub`)           |

Record `llm.request.duration` inside OkHttp instrumentation callback or a thin wrapper around `LlmClient.complete`.

## Environment variables

| Variable                      | Default                 | Purpose                                             |
|-------------------------------|-------------------------|-----------------------------------------------------|
| `OTEL_SERVICE_NAME`           | `ai-temporal-worker`    | `service.name` resource attribute                   |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `http://localhost:4317` | OTLP gRPC endpoint (host worker → Docker collector) |
| `METRICS_PORT`                | `9464`                  | Prometheus scrape bind port                         |
| `OTEL_TRACES_EXPORTER`        | `otlp`                  | Allow `none` for unit tests                         |

## Testing strategy

| Test                                                            | Proves                                  |
|-----------------------------------------------------------------|-----------------------------------------|
| `TemporalTracingInterceptorTest` + `InMemorySpanExporter`       | OTEL-01 attributes on activity span     |
| `MetricsTest` or interceptor test asserting `MetricReader`      | OTEL-02 histogram/counter recorded      |
| `LlmClientTest` with mock server + span exporter                | OTEL-04 child span + propagation header |
| Manual: `./gradlew :worker:run` + `curl localhost:9464/metrics` | Prometheus exposition                   |
| Manual: starter `ping` → JSON log line with `trace_id`          | OTEL-03                                 |

Use `@BeforeEach` / `@AfterEach` to register/shutdown test `OpenTelemetry` instances; avoid polluting global SDK in parallel tests.

## Pitfalls (from `.planning/research/PITFALLS.md`)

- Do not put `workflow_id` on metric labels.
- Do not log full LLM request bodies in spans/logs.
- Span names: `activity.<activityType>`, `llm.complete` — not raw URLs.
- Shutdown order: stop Prometheus server → flush OTLP → shutdown SDK (hook in `WorkerMain`).

## Out of scope (Phase 4)

- Jaeger UI verification (BACK-01, Phase 5)
- Grafana dashboards / Loki pipeline tuning (Phases 6–7)
- Workflow-level spans beyond completion counter (optional stretch; activity spans satisfy OTEL-01)

## References

- `docs/adr/0002-kotlin-worker-opentelemetry.md`
- `docs/adr/0004-workflow-trace-correlation.md`
- [ActivityInboundCallsInterceptor (Temporal SDK 1.25.2)](https://www.javadoc.io/doc/io.temporal/temporal-sdk/1.25.2/io/temporal/common/interceptors/ActivityInboundCallsInterceptor.html)
- [Temporal Code Exchange — native OpenTelemetry](https://www.temporal.io/code-exchange/native-opentelemetry-usage) (patterns; Java/Kotlin custom interceptors)
