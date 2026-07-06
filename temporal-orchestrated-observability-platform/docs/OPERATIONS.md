# Operations Guide

**Audience:** On-call engineers and platform operators  
**Environment:** Local Docker Compose (`deploy/docker-compose.yml`)

## Quick links

| System | URL (default) | Purpose |
|--------|---------------|---------|
| Temporal UI | http://localhost:8080 | Workflow runs, failures, retries, terminate/reset |
| Grafana | http://localhost:3000 | Dashboards, Explore (Loki/Tempo/Prometheus) |
| Jaeger (dev) | http://localhost:16686 | Trace UI (dual-export with Tempo) |
| Prometheus | http://localhost:9090 | Targets, alerts, rules |

## Grafana dashboards

| Dashboard | URL |
|-----------|-----|
| [Workflow Overview](http://localhost:3000/d/workflow-overview) | Completion rate, error ratio |
| [Activity Latency](http://localhost:3000/d/activity-latency) | Activity p95 by type |
| [LLM Proxy](http://localhost:3000/d/llm-proxy) | Stub LLM latency |

Panel details: [docs/GRAFANA.md](GRAFANA.md)

## Correlation cheat sheet

| Signal | Key field | Example |
|--------|-----------|---------|
| Logs (Loki) | `workflow_id`, `trace_id` | `{service_name="ai-temporal-worker"} \| json \| workflow_id="<id>"` |
| Traces (Tempo) | Trace ID from logs | Grafana log line → **View Trace** |
| Traces (Jaeger) | Service `ai-temporal-worker` | Jaeger UI → Search |
| Metrics | `workflow_type`, `activity_type` | `workflow_completed_total{status="error"}` |
| Temporal | Workflow ID | Temporal UI → Workflows |

**End-to-end drill-down:** Temporal UI (failing run) → copy `workflow_id` → Loki filter → open trace from `trace_id` → Activity Latency panel for same `activity_type`.

## Example alert rules (OPS-02)

File: `deploy/prometheus/alerts.yml` (loaded by Prometheus)

| Alert | Condition |
|-------|-----------|
| `HighActivityFailureRate` | >25% `workflow_completed` errors over 5m |
| `AiWorkerTargetDown` | `up{job="ai-worker"} == 0` for 2m |
| `OtelExporterSendFailures` | Collector `otelcol_exporter_send_failed_spans` increases |
| `OtelCollectorTargetDown` | `up{job="otel-collector"} == 0` |

Validate rules:

```bash
docker run --rm --entrypoint promtool \
  -v "$PWD/deploy/prometheus:/etc/prometheus:ro" \
  prom/prometheus:v2.55.1 \
  check rules /etc/prometheus/alerts.yml
```

View firing alerts: http://localhost:9090/alerts

---

## Runbook 1: Stuck workflow

### Symptoms

- Workflow stays **Running** beyond expected duration.
- Activities show repeated attempts in Temporal UI Event History.
- Grafana **Activity Latency** shows sustained activity traffic without completion.

### Detection

1. **Temporal UI** → http://localhost:8080 → namespace `default` → filter by workflow type or ID.
2. **Loki** (Grafana Explore):

   ```logql
   {service_name="ai-temporal-worker"} | json | workflow_id="<workflow-id>"
   ```

3. **Metrics:**

   ```promql
   sum(rate(activity_duration_seconds_count[5m])) by (activity_type, status)
   ```

### Remediation

1. Open the run in Temporal UI → **Event History** — identify failing `activity_type` and error message.
2. If intentional retry demo (`SIMULATE_TOOL_FAILURE=true`), unset env and restart worker:

   ```bash
   unset SIMULATE_TOOL_FAILURE
   ./gradlew :worker:run
   ```

3. **Terminate** a stuck run (CLI — requires `temporal` CLI or use UI):

   - UI: Workflow → **Terminate** (reason: operator intervention).
   - Or reset for a clean replay after fixing the worker/dependency.

4. Re-run with starter after fix:

   ```bash
   ./gradlew :starter:run --args="agent <goal>"
   ```

5. Confirm recovery: workflow **Completed** in UI; `workflow_completed_total{status="ok"}` increases in Prometheus.

---

## Runbook 2: Missing traces

### Symptoms

- JSON logs contain `trace_id`, but **Tempo** / **Jaeger** search returns no trace.
- Grafana **View Trace** link fails or shows empty trace.

### Detection

1. Note `trace_id` from Loki log line.
2. **Tempo** (Grafana Explore → Tempo → Search by trace ID).
3. **Jaeger:** http://localhost:16686 → Service `ai-temporal-worker`.
4. **Collector logs:**

   ```bash
   docker logs temporal-obs-platform-otel-collector-1 --tail 50
   ```

5. **Metric:** `otelcol_exporter_send_failed_spans` in Prometheus (alert `OtelExporterSendFailures`).

### Remediation

1. Confirm worker exports OTLP:

   ```bash
   export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
   ./gradlew :worker:run
   ```

2. Verify collector and backends are healthy:

   ```bash
   curl -fsS http://localhost:13133/          # collector health
   curl -fsS http://localhost:3200/ready      # tempo
   curl -fsS http://localhost:16686/api/services  # jaeger
   ```

3. Recreate collector after config changes:

   ```bash
   docker compose -f deploy/docker-compose.yml up -d --force-recreate otel-collector
   ```

4. Run a short workflow and re-query trace ID:

   ```bash
   ./gradlew :starter:run --args="ping"
   ```

5. If still missing, check `OTEL_TRACES_EXPORTER` is not `none` and worker started **after** `OpenTelemetryConfig.init()`.

---

## Runbook 3: Prometheus target down

### Symptoms

- Grafana metric panels empty.
- Alert `AiWorkerTargetDown` firing.
- http://localhost:9090/targets shows `ai-worker` **DOWN**.

### Detection

```bash
curl -s http://localhost:9090/api/v1/targets | python3 -c "
import json,sys
for t in json.load(sys.stdin)['data']['activeTargets']:
    if t['labels'].get('job')=='ai-worker':
        print(t['health'], t['scrapeUrl'])
"
```

Or open http://localhost:9090/targets → job `ai-worker`.

### Remediation

1. Ensure worker is running and listening on **9464**:

   ```bash
   curl -fsS http://localhost:9464/metrics | head
   ./gradlew :worker:run
   ```

2. Free port if stale process holds 9464:

   ```bash
   lsof -ti :9464 | xargs kill -TERM
   ```

3. **Linux Docker:** Prometheus uses `host.docker.internal` (configured with `extra_hosts: host-gateway` in compose). If still DOWN, scrape from host IP in `deploy/prometheus/prometheus.yml`.

4. Reload Prometheus after config edit:

   ```bash
   curl -X POST http://localhost:9090/-/reload
   ```

5. Confirm **UP** and run `./scripts/otel-smoke.sh`.

---

## Runbook 4: Loki / Tempo disk pressure

### Symptoms

- `docker compose` services OOM or restart loops.
- Loki/Tempo queries timeout in Grafana.
- `docker system df` shows high volume usage.

### Detection

```bash
docker system df
docker stats --no-stream tempo loki promtail grafana
curl -w "%{http_code}" -fsS http://localhost:3100/ready   # loki
curl -w "%{http_code}" -fsS http://localhost:3200/ready   # tempo
```

**Loki** — high ingested volume with bad labels (never label `workflow_id`).

### Remediation

1. Prune unused Docker data (destructive — removes unused images/volumes):

   ```bash
   docker system prune -af --volumes
   ```

2. Restart observability stack:

   ```bash
   docker compose -f deploy/docker-compose.yml down
   docker compose -f deploy/docker-compose.yml up -d
   ```

3. Reduce retention (optional):
   - Tempo: `deploy/tempo/tempo.yaml` → `block_retention`
   - Loki: `deploy/loki/loki-config.yml` limits

4. Truncate local worker log if huge:

   ```bash
   : > logs/worker.log
   ```

5. Re-run sample load: `./scripts/lgtm-sample-load.sh`

---

## Smoke scripts

| Script | Purpose |
|--------|---------|
| `./scripts/smoke.sh` | Compose health (Temporal, Grafana, Loki, Jaeger, collector) |
| `./scripts/otel-smoke.sh` | Worker `:9464/metrics` |
| `./scripts/trace-smoke.sh` | Jaeger trace + Prometheus target |
| `./scripts/lgtm-sample-load.sh` | Logs + metrics + dashboards sample data |

---

*Manual runbook validation: `.planning/phases/07-operations/07-VERIFICATION.md`*
