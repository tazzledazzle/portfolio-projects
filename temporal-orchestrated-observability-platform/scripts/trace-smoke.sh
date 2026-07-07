#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

COMPOSE_FILE="${ROOT}/deploy/docker-compose.yml"
SERVICE="${OTEL_SERVICE_NAME:-ai-temporal-worker}"
JAEGER_URL="${JAEGER_URL:-http://localhost:16686}"
PROMETHEUS_URL="${PROMETHEUS_URL:-http://localhost:9090}"
OTEL_EXPORTER_OTLP_ENDPOINT="${OTEL_EXPORTER_OTLP_ENDPOINT:-http://localhost:4317}"
LLM_STUB_URL="${LLM_STUB_URL:-http://localhost:8090}"
MIN_SPANS="${TRACE_SMOKE_MIN_SPANS:-3}"
WORKER_LOG="/tmp/trace-smoke-worker.log"

wait_url() {
  local name="$1"
  local url="$2"
  local attempts="${3:-30}"
  local i=1
  while (( i <= attempts )); do
    if curl -fsS "$url" >/dev/null 2>&1; then
      echo "PASS  ${name}"
      return 0
    fi
    sleep 2
    i=$((i + 1))
  done
  echo "FAIL  ${name} — not ready: ${url}" >&2
  return 1
}

if command -v lsof >/dev/null 2>&1 && lsof -ti :9464 >/dev/null 2>&1; then
  echo "Freeing port 9464 (stale worker metrics listener) ..."
  kill -TERM $(lsof -ti :9464) 2>/dev/null || true
  sleep 2
fi

echo "Starting observability stack (Temporal, Jaeger, collector, Prometheus, LLM stub) ..."
docker compose -f "${COMPOSE_FILE}" up -d jaeger prometheus llm-stub temporal >/dev/null
docker compose -f "${COMPOSE_FILE}" up -d --force-recreate otel-collector >/dev/null

wait_url "Jaeger API" "${JAEGER_URL}/api/services" 45
wait_url "OTel Collector health" "http://localhost:${OTEL_HEALTH_PORT:-13133}/"
wait_url "Prometheus" "${PROMETHEUS_URL}/-/healthy"
wait_url "LLM stub" "${LLM_STUB_URL}/__admin/health"

echo "Starting worker (OTLP → ${OTEL_EXPORTER_OTLP_ENDPOINT}) ..."
export OTEL_EXPORTER_OTLP_ENDPOINT
export LLM_STUB_URL
./gradlew :worker:run --no-daemon >"${WORKER_LOG}" 2>&1 &
WORKER_PID=$!
trap 'kill -TERM "$WORKER_PID" 2>/dev/null || true' EXIT

for _ in $(seq 1 60); do
  if grep -q "Worker started" "${WORKER_LOG}" 2>/dev/null; then
    break
  fi
  if grep -qE ':worker:run FAILED|BindException: Address already in use' "${WORKER_LOG}" 2>/dev/null; then
    echo "FAIL: worker did not start — see ${WORKER_LOG}" >&2
    exit 1
  fi
  sleep 1
done

if ! grep -q "Worker started" "${WORKER_LOG}" 2>/dev/null; then
  echo "FAIL: worker did not start within 60s — see ${WORKER_LOG}" >&2
  exit 1
fi

echo "Running RAG workflow ..."
if command -v timeout >/dev/null 2>&1; then
  timeout 180 ./gradlew :starter:run --args="rag What is observability?" --no-daemon -q 2>&1 | tee /tmp/trace-smoke-starter.log
else
  ./gradlew :starter:run --args="rag What is observability?" --no-daemon -q 2>&1 | tee /tmp/trace-smoke-starter.log
fi

echo "Waiting for traces in Jaeger (service=${SERVICE}) ..."
found=""
for _ in $(seq 1 30); do
  if python3 - "${JAEGER_URL}" "${SERVICE}" "${MIN_SPANS}" <<'PY'
import json
import sys
import urllib.request

jaeger_url, service, min_spans = sys.argv[1:4]
min_spans = int(min_spans)
url = f"{jaeger_url.rstrip('/')}/api/traces?service={service}&limit=10"
with urllib.request.urlopen(url, timeout=5) as resp:
    payload = json.load(resp)
traces = payload.get("data") or []
if not traces:
    sys.exit(1)
max_spans = max(len(t.get("spans") or []) for t in traces)
if max_spans < min_spans:
    sys.exit(2)
print(f"OK: Jaeger trace with {max_spans} spans (required >= {min_spans})")
PY
  then
    found=1
    break
  fi
  sleep 2
done

if [[ -z "${found}" ]]; then
  echo "FAIL: no Jaeger trace for service '${SERVICE}' with >= ${MIN_SPANS} spans" >&2
  echo "Hint: open ${JAEGER_URL} and search service ${SERVICE}" >&2
  exit 1
fi

echo "Checking Prometheus target ai-worker ..."
python3 - "${PROMETHEUS_URL}" <<'PY'
import json
import sys
import urllib.request

prom_url = sys.argv[1].rstrip("/")
with urllib.request.urlopen(f"{prom_url}/api/v1/targets", timeout=5) as resp:
    data = json.load(resp)
targets = data.get("data", {}).get("activeTargets", [])
worker_targets = [t for t in targets if t.get("labels", {}).get("job") == "ai-worker"]
if not worker_targets:
    raise SystemExit("FAIL: no ai-worker scrape target configured")
health = worker_targets[0].get("health")
if health != "up":
    raise SystemExit(
        f"FAIL: ai-worker target health={health!r} (start worker on host :9464; "
        "on Linux set extra_hosts or run worker in Docker)"
    )
print("OK: Prometheus target ai-worker is UP")
PY

echo ""
echo "Trace smoke passed (Jaeger + Prometheus)."
