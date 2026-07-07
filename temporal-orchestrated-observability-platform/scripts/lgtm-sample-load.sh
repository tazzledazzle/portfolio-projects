#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

COMPOSE_FILE="${ROOT}/deploy/docker-compose.yml"
LOG_FILE="${ROOT}/logs/worker.log"
OTEL_EXPORTER_OTLP_ENDPOINT="${OTEL_EXPORTER_OTLP_ENDPOINT:-http://localhost:4317}"
LLM_STUB_URL="${LLM_STUB_URL:-http://localhost:8090}"
WORKER_LOG="/tmp/lgtm-sample-worker.log"

mkdir -p "${ROOT}/logs"
: >"${LOG_FILE}"

if command -v lsof >/dev/null 2>&1 && lsof -ti :9464 >/dev/null 2>&1; then
  echo "Freeing port 9464 ..."
  kill -TERM $(lsof -ti :9464) 2>/dev/null || true
  sleep 2
fi

echo "Starting LGTM stack + Temporal + LLM stub ..."
docker compose -f "${COMPOSE_FILE}" up -d temporal loki promtail tempo prometheus grafana otel-collector jaeger llm-stub >/dev/null
docker compose -f "${COMPOSE_FILE}" up -d --force-recreate otel-collector >/dev/null

echo "Waiting for Grafana/Loki/Tempo ..."
for url in \
  "http://localhost:${GRAFANA_PORT:-3000}/api/health" \
  "http://localhost:${LOKI_PORT:-3100}/ready" \
  "http://localhost:${TEMPO_HTTP_PORT:-3200}/ready"; do
  for _ in $(seq 1 30); do
    curl -fsS "$url" >/dev/null 2>&1 && break
    sleep 2
  done
done

echo "Starting worker (JSON logs → ${LOG_FILE}) ..."
export OTEL_EXPORTER_OTLP_ENDPOINT LLM_STUB_URL
./gradlew :worker:run --no-daemon 2>&1 | tee -a "${LOG_FILE}" >"${WORKER_LOG}" &
WORKER_PID=$!
trap 'kill -TERM "$WORKER_PID" 2>/dev/null || true' EXIT

for _ in $(seq 1 60); do
  grep -q "Worker started" "${WORKER_LOG}" 2>/dev/null && break
  sleep 1
done

run_workflow() {
  local name="$1"
  shift
  echo ""
  echo "=== ${name} ==="
  ./gradlew :starter:run --args="$*" --no-daemon -q 2>&1 | tee -a "${LOG_FILE}"
}

run_workflow "ping" "ping"
run_workflow "rag" "rag What is observability?"
run_workflow "batch" "batch demo-eval 3"

sleep 5
echo ""
echo "Sample load complete."
echo "  Logs file: ${LOG_FILE}"
echo "  Grafana:   http://localhost:${GRAFANA_PORT:-3000}"
echo "  Loki:      Explore → {service_name=\"ai-temporal-worker\"} |= \"workflow_id\""
echo "  Dashboards: Observability → Workflow Overview | Activity Latency | LLM Proxy"
