#!/usr/bin/env bash
set -euo pipefail

python3 -m pip install -e .
python3 -m changelog_generator.cli --old api_v1.json --new api_v2.json
