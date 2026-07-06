# Phase 4 Verification — Instrumentation (Kotlin + OpenTelemetry)

**Verified:** 2026-05-20  
**Requirements:** OTEL-01 … OTEL-04

## Automated tests

```bash
./gradlew :worker:test :workflows:test
```

| Test | Requirement |
|------|-------------|
| `TemporalTracingInterceptorTest` | OTEL-01 — activity span attributes |
| `LlmClientTracingTest` | OTEL-04 — `traceparent` header + child span |

**Result:** PASS (worker + workflows modules)

## Implementation summary

| Component | Path |
|-----------|------|
| OTel bootstrap | `worker/.../telemetry/OpenTelemetryConfig.kt` |
| Temporal interceptors | `worker/.../telemetry/TemporalTracingInterceptor.kt` |
| Metrics | `worker/.../telemetry/Metrics.kt` |
| JSON logging | `worker/src/main/resources/logback.xml` |
| OkHttp tracing | `worker/.../clients/LlmClient.kt` |
| Worker wiring | `worker/.../WorkerMain.kt` (`WorkerFactoryOptions.setWorkerInterceptors`) |

## OTEL-01 — Activity spans

- Custom `WorkerInterceptor` on `WorkerFactoryOptions` (SDK 1.25.2 has no `WorkerOptions.setWorkerInterceptors`).
- Activity interceptor uses `init(ActivityExecutionContext)` for `ActivityInfo` (not `Activity.getExecutionContext()` in wrapper thread).
- Span attributes: `workflow_id`, `run_id`, `workflow_type`, `activity_type`, `task_queue`.

## OTEL-02 — Prometheus metrics

- `PrometheusHttpServer` on `METRICS_PORT` (default **9464**), path `/metrics`.
- Instruments: `activity.duration`, `workflow.completed`, `llm.request.duration` — bounded labels only (no `workflow_id` on metrics).

## OTEL-03 — JSON logs

- `logstash-logback-encoder` composite JSON to stdout.
- MDC populated in activity interceptor: `trace_id`, `span_id`, Temporal IDs.

Sample log field (from test run, redacted):

```json
{"message":"PingActivity executing","trace_id":"<32-hex>","workflow_id":"test-workflow-id","activity_type":"Ping"}
```

## OTEL-04 — LLM HTTP propagation

- `OkHttpTelemetry.create(openTelemetry).newCallFactory(baseClient)`.
- W3C `traceparent` on outbound stub requests (verified in `LlmClientTracingTest`).

## Manual verification

```bash
docker compose -f deploy/docker-compose.yml up -d temporal otel-collector
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
./gradlew :worker:run
# separate terminal
./gradlew :starter:run --args="ping"
./scripts/otel-smoke.sh
curl -s localhost:9464/metrics | grep -E 'activity_duration|workflow_completed' | head
```

## Deferred to Phase 5

- Jaeger UI trace visibility (BACK-01)
- Prometheus target `UP` in Docker network (BACK-02) — scrape config exists at `deploy/prometheus/prometheus.yml`

## Checklist

- [x] OTEL-01 activity spans with Temporal identifiers
- [x] OTEL-02 Prometheus metrics with bounded labels
- [x] OTEL-03 JSON logs with `trace_id` + Temporal IDs
- [x] OTEL-04 OkHttp child spans + W3C propagation
- [x] Unit tests pass
- [x] `scripts/otel-smoke.sh` added

**Phase 4 complete.** Next: `/gsd-execute-phase 5` (trace/metrics backends).
