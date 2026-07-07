#!/usr/bin/env bash
set -euo pipefail

python3 -m src.reporting.report_writer \
  --graph out/call-graph.json \
  --recency out/recency.json \
  --output out/dead-code-backlog.md
