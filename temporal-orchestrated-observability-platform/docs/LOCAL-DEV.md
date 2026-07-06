# Local Development

## Prerequisites

- **JDK 21** (Temurin recommended)
- **Docker Desktop** with ≥ **8 GB RAM** allocated
- **curl** (for smoke script)

## Start the platform

```bash
cp .env.example .env   # optional — defaults work
docker compose -f deploy/docker-compose.yml up -d
./scripts/smoke.sh
```

## Service URLs

| Service | URL |
|---------|-----|
| Temporal UI | http://localhost:8080 |
| Grafana | http://localhost:3000 (admin / admin from `.env.example`) |
| Prometheus | http://localhost:9090 |
| Loki | http://localhost:3100 |
| OTLP (gRPC) | localhost:4317 |
| Jaeger UI | http://localhost:16686 |
| Worker metrics | http://localhost:9464/metrics |

## Temporal workflow (PingWorkflow)

Start order: **Compose → worker → starter**.

**Terminal 1 — platform**

```bash
docker compose -f deploy/docker-compose.yml up -d
./scripts/smoke.sh
```

**Terminal 2 — worker** (polls task queue `ai-workflows`)

```bash
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
./gradlew :worker:run
# Expect: Worker started — polling ai-workflows at localhost:7233
```

**Terminal 3 — run workflow**

```bash
./gradlew :starter:run --args="ping"
# Expect: workflow_id=ping-<uuid>  result=pong
```

Open **Temporal UI** at http://localhost:8080 → namespace `default` → filter workflow type `PingWorkflow` → status **Completed**.

### Graceful worker shutdown

Send **SIGTERM** to the worker process (`Ctrl+C` in Terminal 2). The worker:

1. Calls `WorkerFactory.shutdown()`
2. Waits up to **30 seconds** for in-flight tasks to finish
3. Shuts down the gRPC connection

Do not `kill -9` during active workflow runs if you want to verify graceful drain behavior.

### Environment

| Variable | Default | Purpose |
|----------|---------|---------|
| `TEMPORAL_HOST` | `localhost:7233` | Temporal frontend gRPC |
| `TEMPORAL_NAMESPACE` | `default` | Namespace for client/worker |

## AI workflows (Phase 3)

Requires **LLM stub** (WireMock) in addition to Temporal:

```bash
docker compose -f deploy/docker-compose.yml up -d temporal temporal-ui llm-stub
export LLM_STUB_URL=http://localhost:8090

./gradlew :worker:run          # terminal 1
./gradlew :starter:run --args="rag What is Temporal?"
./gradlew :starter:run --args="agent Summarize observability"
./gradlew :starter:run --args="batch demo-eval 5"
```

Or run the bundled E2E script:

```bash
./scripts/ai-workflows-e2e.sh
```

**Simulate agent tool retries** (visible in Temporal UI):

```bash
export SIMULATE_TOOL_FAILURE=true
./scripts/ai-workflows-e2e.sh
```

## Build and unit tests

```bash
./gradlew build test
./gradlew test --tests "*Workflow*"
```

## LGTM / Grafana (Phase 6)

Worker JSON logs must reach Loki via **Promtail** (`logs/worker.log` on the host, mounted into the Promtail container).

```bash
mkdir -p logs
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
./scripts/lgtm-sample-load.sh
```

Then open Grafana http://localhost:3000 (admin/admin) → **Observability** folder dashboards. See [docs/GRAFANA.md](GRAFANA.md).

**Loki Explore:**

```logql
{service_name="ai-temporal-worker"} |= "workflow_id"
```

Click **View Trace** on a log line to open the trace in Tempo.

## Traces and metrics (Phase 5)

Collector exports OTLP traces to **Jaeger** (`deploy/otel-collector/config.yaml`). Prometheus scrapes the worker at `host.docker.internal:9464` when the worker runs on the host.

**End-to-end trace smoke** (starts stack, worker, RAG workflow, checks Jaeger + Prometheus):

```bash
./scripts/trace-smoke.sh
```

**Manual checks**

- Jaeger: http://localhost:16686 → Search → Service `ai-temporal-worker`
- Prometheus: http://localhost:9090/targets → job `ai-worker` should be **UP** while the worker is running

## Troubleshooting

| Symptom | Fix |
|---------|-----|
| `no space left on device` | `docker system prune -af --volumes` then retry |
| Temporal unhealthy | Ensure `BIND_ON_IP=0.0.0.0` and `DB=postgres12` in compose |
| Port 8080 in use | Set `TEMPORAL_UI_PORT` in `.env` |
