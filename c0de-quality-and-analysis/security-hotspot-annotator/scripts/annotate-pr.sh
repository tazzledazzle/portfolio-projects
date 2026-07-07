#!/usr/bin/env bash
set -euo pipefail

python3 -m src.publish.pr_comment_client \
  --semgrep semgrep-results.json \
  --policy config/severity-policy.yml
