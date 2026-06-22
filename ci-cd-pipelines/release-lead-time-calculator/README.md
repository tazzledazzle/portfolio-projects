# Release Lead Time Calculator

## Overview
Release Lead Time Calculator estimates lead time and cycle time from merge and deploy timestamps, exposing summary metrics useful for DORA reporting.

## Key Feature
Its unique feature is direct merge-to-deploy distribution analysis that can be surfaced as API-ready DORA metrics.

## Architecture
- `src/main.py`: Time-series metric calculations.
- `tests/test_smoke.py`: Validates percentile and aggregate calculations.
- `Makefile`: Run/test helpers.

## Use Cases
- Track release performance per service/team.
- Detect deployment bottlenecks.
- Feed dashboards with lead-time trends.

## Usage
```bash
make run
make test
```

## Control Flow
1. Load merged PR and deploy event pairs.
2. Compute per-item lead times.
3. Aggregate mean/p50/p90 metrics.
4. Output metrics payload for downstream APIs.

## Project Structure
```text
release-lead-time-calculator/
  README.md
  Makefile
  src/
    main.py
  tests/
    test_smoke.py
```
