# OTel Demo Stack

**Problem:** Demonstrate end-to-end observability with OpenTelemetry: one service, one worker, traces and metrics flowing to a backend (collector / Grafana or similar).

**Solution:** A minimal "observability in a box" setup: an instrumented API, an instrumented worker, OTLP export to an OpenTelemetry Collector, with W3C trace context propagation. Includes an SRE-facing **OpenTelemetry Verification Guide** and optional CI that checks the OTel health endpoint.

**What this demonstrates:** Designing and documenting an end-to-end observability pipeline and writing SRE verification docs—aligned with Temporal Worker observability, OTel verification guides, and OTLP/Datadog-style workflows.

---

## Contents

- **api:** Small FastAPI service with OTel instrumentation; exposes `/api/health` and `/api/health/otel`.
- **worker:** Minimal Python worker that does a periodic task and emits spans/metrics; propagates W3C trace context.
- **collector:** OpenTelemetry Collector config (OTLP in, export to stdout or Prometheus/OTLP).
- **docker-compose:** Runs api, worker, and collector for local use.
- **Verification:** See [docs/opentelemetry-verification-guide.md](docs/opentelemetry-verification-guide.md) for how to confirm traces and metrics are flowing.

---

## How to run

**With Docker Compose:**

```bash
docker compose up --build
```

- API: http://localhost:8000  
- API health: http://localhost:8000/health  
- OTel health: http://localhost:8000/api/health/otel  

Trigger a request to the API; the worker runs on an interval. Traces and metrics are sent to the collector (see collector logs or configure a backend in `collector/otelcol.yaml`).

**Verification steps:** See [OpenTelemetry Verification Guide](docs/opentelemetry-verification-guide.md).

---

## License

MIT.
