# Phase 5: Trace & Metrics Backends — Context

**Gathered:** 2026-05-19

<domain>

## Phase boundary

Deliver **BACK-01**, **BACK-02**, **BACK-03**:

- OTel Collector receives OTLP from worker; exports traces to **Jaeger**.
- Prometheus scrapes worker `:9464` and collector self-metrics.
- CI validates collector YAML.

Maps diagram edges: SDK → **traces** → Jaeger; SDK → **metrics** → Prometheus.

**Out of scope:** Tempo as primary (Phase 6); Loki pipelines beyond Promtail config stub.

</domain>

<decisions>

- **D-01:** Collector config: `otlp` receiver → `jaeger` exporter + `batch` processor.
- **D-02:** Jaeger all-in-one in Compose (enable profile if disabled in Phase 1).
- **D-03:** Prometheus scrape job `ai-worker` host `host.docker.internal` or service name `worker` if worker containerized later.
- **D-04:** Document local dev: worker on host, collector in Docker — use `OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317`.

</decisions>

<canonical_refs>

- `docs/adr/0003-jaeger-dev-tempo-lgtm.md`
- `.planning/phases/04-instrumentation/04-VERIFICATION.md`

</canonical_refs>
