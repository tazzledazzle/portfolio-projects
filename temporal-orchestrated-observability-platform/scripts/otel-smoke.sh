#!/usr/bin/env bash
set -euo pipefail

METRICS_PORT="${METRICS_PORT:-9464}"
METRICS_URL="http://localhost:${METRICS_PORT}/metrics"

echo "Checking Prometheus metrics at ${METRICS_URL} ..."
if ! curl -sf "${METRICS_URL}" >/tmp/otel-smoke-metrics.txt; then
  echo "FAIL: metrics endpoint unreachable. Start the worker: ./gradlew :worker:run" >&2
  exit 1
fi

if ! grep -q 'activity_duration' /tmp/otel-smoke-metrics.txt; then
  echo "WARN: activity_duration not found yet — run a workflow (e.g. ./gradlew :starter:run --args='ping') then re-run this script." >&2
  exit 1
fi

echo "OK: metrics endpoint exposes activity_duration"
grep -E '^(activity_duration|workflow_completed)' /tmp/otel-smoke-metrics.txt | head -5 || true
