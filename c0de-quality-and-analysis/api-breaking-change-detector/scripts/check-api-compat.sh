#!/usr/bin/env bash
set -euo pipefail

python3 -m src.ci.merge_gate \
  --base specs/main-openapi.yml \
  --head specs/branch-openapi.yml \
  --policy config/compatibility-policy.yml
