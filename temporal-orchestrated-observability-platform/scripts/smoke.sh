#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_FILE="${ROOT}/deploy/docker-compose.yml"

TEMPORAL_UI_PORT="${TEMPORAL_UI_PORT:-8080}"
GRAFANA_PORT="${GRAFANA_PORT:-3000}"
PROMETHEUS_PORT="${PROMETHEUS_PORT:-9090}"
LOKI_PORT="${LOKI_PORT:-3100}"
JAEGER_UI_PORT="${JAEGER_UI_PORT:-16686}"
OTEL_HEALTH_PORT="${OTEL_HEALTH_PORT:-13133}"

MAX_ATTEMPTS="${SMOKE_MAX_ATTEMPTS:-36}"
SLEEP_SECS="${SMOKE_SLEEP_SECS:-5}"

check_url() {
  local name="$1"
  local url="$2"
  local attempt=1

  echo "Checking ${name} at ${url} ..."
  while (( attempt <= MAX_ATTEMPTS )); do
    if curl -fsS "$url" >/dev/null 2>&1; then
      echo "PASS  ${name}"
      return 0
    fi
    echo "  attempt ${attempt}/${MAX_ATTEMPTS} — not ready, sleeping ${SLEEP_SECS}s"
    sleep "${SLEEP_SECS}"
    attempt=$((attempt + 1))
  done

  echo "FAIL  ${name} — timed out after $((MAX_ATTEMPTS * SLEEP_SECS))s"
  return 1
}

if ! command -v curl >/dev/null 2>&1; then
  echo "curl is required" >&2
  exit 1
fi

if ! command -v docker >/dev/null 2>&1; then
  echo "docker is required" >&2
  exit 1
fi

echo "Validating compose file ..."
docker compose -f "${COMPOSE_FILE}" config >/dev/null

echo "Starting platform (docker compose up -d) ..."
docker compose -f "${COMPOSE_FILE}" up -d

check_url "Temporal UI" "http://localhost:${TEMPORAL_UI_PORT}/"
check_url "Grafana" "http://localhost:${GRAFANA_PORT}/api/health"
check_url "Prometheus" "http://localhost:${PROMETHEUS_PORT}/-/healthy"
check_url "Loki" "http://localhost:${LOKI_PORT}/ready"
check_url "Jaeger UI" "http://localhost:${JAEGER_UI_PORT}/"
check_url "OTel Collector" "http://localhost:${OTEL_HEALTH_PORT}/"

echo ""
echo "All smoke checks passed."
