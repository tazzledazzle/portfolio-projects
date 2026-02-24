# OpenTelemetry Verification Guide (SRE)

This guide describes how to verify that OpenTelemetry traces and metrics from the **otel-demo-stack** are being emitted, exported, and (optionally) visible in your backend.

---

## 1. Health check endpoint

The API exposes an OTel health endpoint:

```bash
curl -s http://localhost:8000/api/health/otel | jq .
```

**Expected (healthy):**

- `sdk.initialized`: `true`
- `exporters[].healthy`: `true` (for each configured exporter)
- `testSpan.created` and `testSpan.exported`: `true` (if the app sends a test span on health check)

If any exporter is `healthy: false`, check network connectivity to the OTLP endpoint (e.g. collector address and port).

---

## 2. Application logs

Check API and worker logs for OTel initialization:

- Look for messages like: `OTel SDK started`, `OTLP exporter configured`, or `TracerProvider set`.
- Errors such as `connection refused` or `timeout` indicate the collector (or OTLP backend) is unreachable.

---

## 3. Verify export endpoint is reachable

From the same network as the app (e.g. from the API container):

```bash
# OTLP HTTP (common)
curl -s -o /dev/null -w "%{http_code}" http://otel-collector:4318/v1/traces
# Expect 200 or 404 (some collectors return 404 for GET; POST is used for export)
```

Ensure the collector service name and port match what the API and worker use in `OTEL_EXPORTER_OTLP_ENDPOINT`.

---

## 4. Generate traffic and confirm data

1. **Traces:** Call the API (e.g. `curl http://localhost:8000/health` and any other instrumented route). The worker also creates spans on its interval. In your backend (Grafana Tempo, Jaeger, or Datadog), search by service name (`otel-demo-api`, `otel-demo-worker`) and time range; confirm spans appear and, if applicable, share the same trace ID across API and worker.
2. **Metrics:** Query your metrics backend (Prometheus or Datadog) for the service name or `otel_*` / SDK metric names. Confirm recent data.

---

## 5. CI verification (optional)

The GitHub Actions workflow can run the stack (or only the API), call `/api/health/otel`, and assert that `sdk.initialized` is true and exporters are healthy. This catches misconfiguration before merge.

---

## Troubleshooting

| Symptom | Possible cause | Action |
|--------|----------------|--------|
| Health endpoint returns 500 or SDK not initialized | Env vars missing or wrong (e.g. OTEL_SERVICE_NAME, OTEL_EXPORTER_OTLP_ENDPOINT) | Check env in docker-compose or deployment |
| Exporters unhealthy | Collector not running or wrong host/port | Verify collector is up and reachable from API/worker |
| Spans in app but not in backend | Export failure (network, TLS, batch) | Check collector logs and SDK exporter logs |
| No trace correlation between API and worker | W3C context not propagated | Ensure both use W3C Trace Context propagator and same OTLP endpoint |
