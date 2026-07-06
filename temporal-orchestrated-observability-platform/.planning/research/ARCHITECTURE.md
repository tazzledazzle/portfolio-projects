# Research — Architecture

**Researched:** 2026-05-19

## Major components (diagram → implementation)

| Diagram box | Implementation unit | Phase |
|-------------|---------------------|-------|
| AI Workflow 1..N | `workflows/` Kotlin interfaces + `starter/` CLI | 3 |
| Temporal Engine | `deploy/docker-compose` Temporal service | 2 |
| OpenTelemetry SDK | `worker/telemetry/` module | 4 |
| Jaeger | Compose service + collector exporter | 5 |
| Prometheus | Compose + scrape config | 5 |
| Loki / Tempo / Grafana | Compose LGTM profile | 6 |
| Operations / On-Call | `docs/OPERATIONS.md` + alerts | 7 |

## Data flow (happy path)

```
Starter → Temporal (start workflow) → Worker (workflow task)
  → Activities (LLM/retrieve stubs) → OTel spans/metrics/logs
  → Collector → Jaeger/Tempo + Prometheus + Loki
  → Grafana dashboards → Operator
```

## Suggested build order

1. Compose + Gradle skeleton (no business logic)
2. Temporal ping workflow
3. Three AI workflows with stubs
4. OTel on worker (console → OTLP)
5. Jaeger + Prometheus wiring
6. Loki + Tempo + Grafana provisioning
7. Runbooks + verification

## Boundaries

- **Workflows** never call HTTP/LLM directly (Temporal determinism).
- **Activities** own all side effects and instrumentation.
- **Grafana** never queries Temporal directly—use metrics/traces/logs signals only.
- **Jaeger** is dev aid; **Tempo** is Grafana source of truth after Phase 6.

## Integration points

| From | To | Protocol |
|------|-----|----------|
| Starter | Temporal | gRPC (frontend) |
| Worker | Temporal | gRPC (worker) |
| Worker | LLM stub | HTTP + traceparent |
| Worker | OTel Collector | OTLP gRPC :4317 |
| Prometheus | Worker | HTTP scrape :9464 |
| Promtail | Loki | HTTP push |
| Grafana | LGTM stores | Native datasources |
