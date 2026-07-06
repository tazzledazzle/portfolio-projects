# Phase 4 Plan Check

**Checked:** 2026-05-20  
**Phase goal:** Worker emits correlated traces, metrics, and logs for all activities (OTEL-01 … OTEL-04).

## Requirement coverage

| Requirement | Plan | Tasks | Verdict |
|-------------|------|-------|---------|
| OTEL-01 Activity spans + Temporal IDs | 04-01 | Task 4, 5 | PASS |
| OTEL-02 Prometheus histograms/counters | 04-01 | Task 3, 4, 5 | PASS |
| OTEL-03 JSON logs + trace_id | 04-02 | Task 1 | PASS |
| OTEL-04 OkHttp child spans + W3C | 04-02 | Task 2, 3 | PASS |

## Goal-backward truths

| Truth | Planned proof | Verdict |
|-------|---------------|---------|
| Spans carry Temporal identifiers | `TemporalTracingInterceptorTest` | PASS |
| `:9464/metrics` exposes bounded labels | Manual curl + Metrics instruments without `workflow_id` label | PASS |
| JSON logs include `trace_id` | logback JSON + MDC in interceptor | PASS |
| LLM HTTP child span + propagation | `LlmClientTracingTest` | PASS |

## ADR / pitfall alignment

| Rule | Plan enforcement | Verdict |
|------|------------------|---------|
| ADR-004: no `workflow_id` on metrics | Metrics.kt attribute list | PASS |
| ADR-004: IDs on spans/logs | Interceptor attributes + MDC | PASS |
| PITFALLS: no prompt text in telemetry | Task notes in 04-02 | PASS |
| D-05 temporal-opentelemetry | 04-RESEARCH: custom interceptor | PASS |

## Dependency order

```text
03-ai-workflows → 04-01 (SDK + interceptor + metrics) → 04-02 (logs + OkHttp + smoke)
05-backends (Jaeger scrape proof) → depends on 04 complete
```

## Gaps / risks

| Risk | Mitigation in plan |
|------|-------------------|
| `opentelemetry-exporter-prometheus` alpha API drift | Pin BOM 1.43; compile in Task 1 |
| Host worker → Docker OTLP networking | Document `localhost:4317`; Phase 5 adds troubleshooting |
| Workflow counter without workflow spans | Workflow inbound interceptor increments counter only |

## Overall

**READY FOR EXECUTION** — run `/gsd-execute-phase 4` (executor should run 04-01 then 04-02).
