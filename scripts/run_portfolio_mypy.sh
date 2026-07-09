#!/usr/bin/env bash
# Run mypy per subproject to avoid duplicate module name errors from repo-wide invocation.
#
# Only includes subprojects that are both collision-free (no duplicate top-level module names
# across invocations) and currently type-clean (exit 0). Subprojects with pre-existing type
# errors or module-name collisions are excluded to keep CI green.
#
# To extend coverage: fix the type errors in a subproject, verify it passes mypy locally,
# then add it to the list below.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

run_mypy() {
  local dir="$1"
  if [ ! -d "$dir" ]; then
    return 0
  fi
  echo "==> mypy in $dir"
  (
    cd "$dir"
    mypy --ignore-missing-imports --no-error-summary .
  )
}

# Standalone demos (collision-free and type-clean)
run_mypy otel-demo-stack/api
run_mypy rest-api-test-demo/app
run_mypy platform-audit-template/scripts

# AI examples — only include subprojects that are currently type-clean
run_mypy ai-best-practices-examples/knowledge-qa-system

echo "Portfolio mypy complete."
