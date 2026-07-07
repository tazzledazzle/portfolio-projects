#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

echo "Ensuring Temporal is up ..."
docker compose -f deploy/docker-compose.yml up -d temporal temporal-ui >/dev/null

echo "Starting worker ..."
./gradlew :worker:run --no-daemon >/tmp/ping-e2e-worker.log 2>&1 &
WORKER_PID=$!
trap 'kill -TERM "$WORKER_PID" 2>/dev/null || true' EXIT

for _ in $(seq 1 30); do
  if grep -q "Worker started" /tmp/ping-e2e-worker.log 2>/dev/null; then
    break
  fi
  sleep 1
done

if ! grep -q "Worker started" /tmp/ping-e2e-worker.log; then
  echo "FAIL: worker did not start" >&2
  tail -20 /tmp/ping-e2e-worker.log >&2
  exit 1
fi

echo "Running starter ping ..."
OUTPUT="$(./gradlew :starter:run --args="ping" --no-daemon -q 2>&1)"
echo "$OUTPUT"

echo "$OUTPUT" | grep -q "result=pong" || { echo "FAIL: expected result=pong"; exit 1; }
echo "$OUTPUT" | grep -q "workflow_id=ping-" || { echo "FAIL: expected workflow_id"; exit 1; }

echo "PASS  PingWorkflow E2E"
