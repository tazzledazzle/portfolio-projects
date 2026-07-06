# Phase 6 Verification — LGTM / Grafana

**Verified:** 2026-05-20  
**Requirements:** LGTM-01 … LGTM-04

## Changes

| Artifact | Purpose |
|----------|---------|
| `deploy/otel-collector/config.yaml` | Dual-export traces to **Tempo** + **Jaeger** |
| `deploy/promtail/config.yml` | JSON worker log scrape → Loki (`service_name`, `level` labels only) |
| `deploy/docker-compose.yml` | Promtail mounts `logs/`; Grafana mounts dashboards |
| `deploy/grafana/provisioning/datasources/datasources.yml` | Stable UIDs; Loki `trace_id` → Tempo derived field |
| `deploy/grafana/provisioning/dashboards/dashboards.yml` | File provisioning |
| `deploy/grafana/dashboards/*.json` | Workflow Overview, Activity Latency, LLM Proxy |
| `scripts/lgtm-sample-load.sh` | Sample workflows + log tee |
| `docs/GRAFANA.md`, `docs/OPERATIONS.md` | Dashboard docs + quick links (LGTM-04) |

## Automated validation

```bash
docker run --rm \
  -v "$PWD/deploy/otel-collector/config.yaml:/etc/otelcol-contrib/config.yaml:ro" \
  otel/opentelemetry-collector-contrib:0.114.0 \
  validate --config=/etc/otelcol-contrib/config.yaml

docker compose -f deploy/docker-compose.yml config
./gradlew test
```

**Result:** PASS (config validate + unit tests)

## LGTM-01 — Loki logs

**Config:** Promtail reads `logs/worker.log` (host) via volume `../logs:/var/log/worker`.

**Query (Grafana Explore → Loki):**

```logql
{service_name="ai-temporal-worker"} |= "workflow_id"
```

**Load:** `./scripts/lgtm-sample-load.sh`

## LGTM-02 — Tempo + log correlation

- Collector exports OTLP to `tempo:4317`.
- Loki derived field **View Trace** links `trace_id` → Tempo datasource (`uid: tempo`).
- Tempo `tracesToLogs` links back to Loki.

**Manual:** After sample load, open a log line in Explore and click **View Trace**.

## LGTM-03 — Dashboards

Provisioned under folder **Observability**:

| File | UID |
|------|-----|
| `workflow-overview.json` | `workflow-overview` |
| `activity-latency.json` | `activity-latency` |
| `llm-proxy.json` | `llm-proxy` |

Metrics: `workflow_completed_total`, `activity_duration_seconds_*`, `llm_request_duration_seconds_*`.

## LGTM-04 — Documentation

- [docs/OPERATIONS.md](../../docs/OPERATIONS.md) — dashboard URLs table
- [docs/GRAFANA.md](../../docs/GRAFANA.md) — panel expectations

## Checklist

- [x] LGTM-01 Promtail → Loki JSON pipeline
- [x] LGTM-02 Tempo ingest + derived fields
- [x] LGTM-03 Three provisioned dashboards
- [x] LGTM-04 Operations + Grafana docs
- [x] Collector config validates with dual exporters

**Phase 6 complete.** Next: `/gsd-execute-phase 7` (Operations: alerts + runbook validation).
