#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

export LLM_STUB_URL="${LLM_STUB_URL:-http://localhost:8090}"
export SIMULATE_TOOL_FAILURE="${SIMULATE_TOOL_FAILURE:-false}"

echo "Starting Temporal + LLM stub ..."
docker compose -f deploy/docker-compose.yml up -d temporal temporal-ui llm-stub >/dev/null

echo "Waiting for Temporal ..."
for _ in $(seq 1 60); do
  if curl -fsS "http://localhost:${TEMPORAL_UI_PORT:-8080}/" >/dev/null 2>&1; then
    break
  fi
  sleep 2
done

echo "Waiting for LLM stub ..."
for _ in $(seq 1 30); do
  if curl -fsS "${LLM_STUB_URL}/__admin/health" >/dev/null 2>&1; then
    break
  fi
  sleep 2
done

echo "Starting worker ..."
./gradlew :worker:run --no-daemon >/tmp/ai-workflows-worker.log 2>&1 &
WORKER_PID=$!
trap 'kill -TERM "$WORKER_PID" 2>/dev/null || true' EXIT

for _ in $(seq 1 45); do
  if grep -q "Worker started" /tmp/ai-workflows-worker.log 2>/dev/null; then
    break
  fi
  sleep 1
done

run_cmd() {
  local name="$1"
  shift
  echo ""
  echo "=== $name ==="
  local out
  out="$(./gradlew :starter:run --args="$*" --no-daemon -q 2>&1)"
  echo "$out"
  echo "$out" | grep -q "workflow_id=" || { echo "FAIL: missing workflow_id"; exit 1; }
}

run_cmd "RAG" "rag What is observability?"
run_cmd "Agent" "agent Summarize the platform"
run_cmd "Batch" "batch demo-eval 5"

if [[ "${SIMULATE_TOOL_FAILURE}" == "true" ]]; then
  echo ""
  echo "NOTE: SIMULATE_TOOL_FAILURE=true — check Temporal UI for AgentToolsWorkflow retries"
fi

echo ""
echo "PASS  All AI workflow starter commands completed"
