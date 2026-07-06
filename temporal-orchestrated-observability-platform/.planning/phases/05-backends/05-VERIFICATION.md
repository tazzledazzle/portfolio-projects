# Phase 5 Verification — Trace & Metrics Backends

**Verified:** 2026-05-20  
**Requirements:** BACK-01 … BACK-03

## Changes

| Artifact | Change |
|----------|--------|
| `deploy/otel-collector/config.yaml` | OTLP receiver → `otlp/jaeger` exporter (replaces Tempo for dev traces) |
| `deploy/docker-compose.yml` | Jaeger in default stack; collector `depends_on: jaeger` |
| `deploy/prometheus/prometheus.yml` | Scrape `ai-worker` + `otel-collector:8888` |
| `scripts/trace-smoke.sh` | RAG run + Jaeger API + Prometheus target check |
| `scripts/smoke.sh` | Jaeger UI health check |
| `.github/workflows/ci.yml` | OTel collector validate (pre-existing) |

## BACK-03 — Collector validate

```bash
docker run --rm \
  -v "$PWD/deploy/otel-collector/config.yaml:/etc/otelcol-contrib/config.yaml:ro" \
  otel/opentelemetry-collector-contrib:0.114.0 \
  validate --config=/etc/otelcol-contrib/config.yaml
```

**Result:** PASS

```bash
docker compose -f deploy/docker-compose.yml config
```

**Result:** PASS

## BACK-01 — Jaeger traces

**Automated:** `./scripts/trace-smoke.sh`

- Starts `jaeger`, `otel-collector`, `temporal`, `prometheus`, `llm-stub`
- Worker exports OTLP to `localhost:4317`
- Runs `rag` workflow
- Queries `http://localhost:16686/api/traces?service=ai-temporal-worker` — expects ≥3 spans

**Manual:** Jaeger UI → Service `ai-temporal-worker` after a RAG run.

## BACK-02 — Prometheus scrape

- Job `ai-worker` → `host.docker.internal:9464/metrics` (worker on host)
- Job `otel-collector` → `otel-collector:8888`

`trace-smoke.sh` asserts `ai-worker` target **UP** while worker runs.

**Linux note:** `host.docker.internal` may need `extra_hosts` in Compose if scrape stays down on native Linux.

## Checklist

- [x] BACK-01 Jaeger receives OTLP traces from collector
- [x] BACK-02 Prometheus scrape config for worker + collector
- [x] BACK-03 CI/local `otelcol validate`
- [x] `scripts/trace-smoke.sh` added
- [x] Docs updated (`docs/LOCAL-DEV.md`)

**Phase 5 complete.** Next: `/gsd-execute-phase 6` (LGTM / Grafana).
