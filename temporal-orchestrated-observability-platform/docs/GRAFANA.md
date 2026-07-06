# Grafana — LGTM Dashboards

**Stack:** Prometheus (metrics) · Loki (logs) · Tempo (traces) · Grafana 11

## Access

| Item | Default |
|------|---------|
| URL | http://localhost:3000 |
| User | `admin` (from `.env` / `GRAFANA_ADMIN_USER`) |
| Password | `admin` (from `GRAFANA_ADMIN_PASSWORD`) |

## Provisioned dashboards

Folder **Observability**:

| Dashboard | UID | Path in UI |
|-----------|-----|------------|
| Workflow Overview | `workflow-overview` | Dashboards → Observability → Workflow Overview |
| Activity Latency | `activity-latency` | Dashboards → Observability → Activity Latency |
| LLM Proxy | `llm-proxy` | Dashboards → Observability → LLM Proxy |

Direct links (after login):

- http://localhost:3000/d/workflow-overview
- http://localhost:3000/d/activity-latency
- http://localhost:3000/d/llm-proxy

## Panel expectations (after sample load)

Run `./scripts/lgtm-sample-load.sh` with the worker tee-ing JSON logs to `logs/worker.log`.

### Workflow Overview

- **Workflow completion rate** — non-zero lines per `workflow_type` / `status` after `ping`, `rag`, or `batch`.
- **Workflow error ratio** — near 0 in happy path; rises if activities fail.
- **Workflow completions (1h)** — stat ≥ 3 after sample script.

### Activity Latency

- **Activity duration percentiles** — p50/p95/p99 per `activity_type` (e.g. `Ping`, `embedQuery`, `llmComplete`).
- **Activity execution rate** — increases during workflow runs.

### LLM Proxy

- **LLM request duration** — p50/p95 for `status=ok` after RAG runs (WireMock stub).
- **LLM request rate** — spikes during `llmComplete` activities.
- **LLM stub outcomes** — OK count ≥ 1 after RAG.

## Explore — logs and traces

### Loki (LGTM-01)

```logql
{service_name="ai-temporal-worker"} |= "workflow_id"
```

Filter by workflow:

```logql
{service_name="ai-temporal-worker"} | json | workflow_id="<your-workflow-id>"
```

### Log → trace (LGTM-02)

1. Open a log line with `trace_id` in JSON.
2. Click **View Trace** (derived field → Tempo datasource `tempo`).
3. Or Explore → Tempo → Search → paste trace ID from logs.

Traces are exported to **Tempo** and **Jaeger** via the OTel Collector (`deploy/otel-collector/config.yaml`).

## Datasources (provisioned)

| Name | UID | URL (in Compose network) |
|------|-----|---------------------------|
| Prometheus | `prometheus` | http://prometheus:9090 |
| Loki | `loki` | http://loki:3100 |
| Tempo | `tempo` | http://tempo:3200 |

Config: `deploy/grafana/provisioning/datasources/datasources.yml`
