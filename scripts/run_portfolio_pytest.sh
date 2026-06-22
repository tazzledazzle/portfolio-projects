#!/usr/bin/env bash
# Run pytest per subproject to avoid ImportPathMismatchError from repo-wide collection.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

run_pytest() {
  local dir="$1"
  shift
  if [ ! -d "$dir" ]; then
    return 0
  fi
  if [ ! -d "$dir/tests" ] && [ ! -d "$dir/test" ]; then
    return 0
  fi
  echo "==> pytest in $dir"
  (
    cd "$dir"
    if [ -f pyproject.toml ]; then
      python3 -m pip install -q -e .
    fi
    if ls tests/test_*.py >/dev/null 2>&1 && grep -q 'from src\.' tests/test_*.py 2>/dev/null; then
      export PYTHONPATH="."
    else
      export PYTHONPATH="src${PYTHONPATH:+:$PYTHONPATH}"
    fi
    python3 -m pytest -q "$@"
  )
}

# Standalone demos
echo "==> pytest in projgen"
(cd projgen && python3 -m pytest -q src/tests)
run_pytest online-bookstore
run_pytest rest-api-test-demo
run_pytest workflow-api-demo/api
run_pytest otel-demo-stack/api
run_pytest platform-audit-template/scripts

# AI examples
for dir in ai-best-practices-examples/*/; do
  run_pytest "$dir"
done

# Code quality tools
for dir in c0de-quality-and-analysis/*/; do
  case "$dir" in
    */kotlin-custom-detekt-rules-library/) continue ;;
  esac
  run_pytest "$dir"
done

# CI/CD pipeline scaffolds
for dir in ci-cd-pipelines/*/; do
  run_pytest "$dir"
done

# DevEx utilities
for dir in dev-ex/*/; do
  case "$dir" in
    */tooling-adoption-tracker/) continue ;;
  esac
  run_pytest "$dir"
done

echo "Portfolio pytest matrix complete."
