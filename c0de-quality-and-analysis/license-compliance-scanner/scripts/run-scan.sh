#!/usr/bin/env bash
set -euo pipefail

python3 -m src.policy.policy_engine \
  --deps build/dependency-tree.json \
  --policy config/license-policy.yml \
  --sbom out/sbom.json
