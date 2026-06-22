#!/usr/bin/env bash
set -euo pipefail

python3 -m pip install -e .
uvicorn app.main:app --reload --app-dir src
