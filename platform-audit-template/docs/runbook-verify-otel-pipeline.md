# Runbook: Verify OpenTelemetry Pipeline

**Purpose:** Confirm that OpenTelemetry traces and metrics are being emitted, exported, and visible in your backend (e.g. collector, Grafana Tempo/Prometheus, or Datadog).

**When to use:** After deploying a new service, changing OTel config, or when investigating "no traces/metrics" reports.

**Prerequisites:** Access to the service or pod, and to the OTel backend (or collector) if applicable.

---

## 1. Verify SDK / instrumentation is loaded

- Check application logs for OTel initialization (e.g. "OTel SDK started", "OTLP exporter configured").
- If the app exposes an OTel health endpoint (e.g. `/api/health/otel`), call it and confirm `initialized: true` and exporters healthy.

```bash
curl -s http://localhost:8080/api/health/otel | jq .
```

---

## 2. Verify export endpoint is reachable

- From the same network as the app (e.g. from the pod or same host), ensure the OTLP endpoint is reachable:

```bash
# Example: OTLP HTTP
curl -s -o /dev/null -w "%{http_code}" http://otel-collector:4318/v1/traces

# Example: OTLP gRPC (may require grpcurl)
grpcurl -plaintext otel-collector:4317 list
```

- If using an in-cluster collector, use the in-cluster DNS name (e.g. `http://otel-collector.monitoring.svc.cluster.local:4318`).

---

## 3. Generate a test span

- Trigger a request or code path that is instrumented (e.g. HTTP request to the service).
- Optionally use a dedicated test endpoint that creates a span and returns the trace ID for correlation.

---

## 4. Confirm data in backend

- **Traces:** In your APM/tracing UI (Datadog, Tempo, Jaeger), search by service name and time range; confirm the trace appears and has the expected spans.
- **Metrics:** In your metrics UI (Prometheus, Datadog), query for the service (e.g. `otel_*` or service-specific metrics) and confirm recent data.

---

## 5. If verification fails

| Symptom | Possible cause | Action |
|--------|----------------|--------|
| No OTel logs | Instrumentation not loaded or wrong config | Check env vars (OTEL_EXPORTER_OTLP_ENDPOINT, OTEL_SERVICE_NAME), dependency injection order |
| Health endpoint unhealthy | Exporter unreachable or misconfigured | Check network, TLS, collector config |
| Spans in app but not in backend | Export failure (network, auth, batch size) | Check collector logs, exporter errors in SDK logs |
| Partial traces | Sampling or context propagation broken | Check W3C propagation, sampler config |

---

## Escalation

- Document trace ID and time range when escalating.
- Attach relevant log snippets (no secrets).
- If using Datadog/OTel Collector, include collector and agent logs if applicable.
