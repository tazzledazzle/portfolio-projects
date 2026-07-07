#!/usr/bin/env bash

set -euo pipefail

HOST="${AIVG_SANITY_HOST:-127.0.0.1}"
PORT="${AIVG_SANITY_PORT:-7860}"
URL="http://${HOST}:${PORT}"
STARTUP_TIMEOUT_SECONDS="${AIVG_SANITY_TIMEOUT_SECONDS:-30}"

cleanup() {
  if [[ -n "${APP_PID:-}" ]] && kill -0 "${APP_PID}" 2>/dev/null; then
    kill "${APP_PID}" 2>/dev/null || true
    wait "${APP_PID}" 2>/dev/null || true
  fi
}

trap cleanup EXIT

python3 -m ai_image_video_generator.app > /tmp/aivg_sanity.log 2>&1 &
APP_PID=$!

echo "Started app pid=${APP_PID}; waiting for ${URL}"

for ((i = 1; i <= STARTUP_TIMEOUT_SECONDS; i++)); do
  if curl -sSf "${URL}" > /dev/null; then
    echo "Sanity check passed: Gradio reachable at ${URL}"
    exit 0
  fi
  sleep 1
done

echo "Sanity check failed: Gradio did not become reachable at ${URL} within ${STARTUP_TIMEOUT_SECONDS}s"
echo "---- app log tail ----"
python3 - <<'PY'
from pathlib import Path

log = Path("/tmp/aivg_sanity.log")
if not log.exists():
    print("No log file found.")
else:
    lines = log.read_text(encoding="utf-8", errors="replace").splitlines()
    for line in lines[-40:]:
        print(line)
PY
exit 1
