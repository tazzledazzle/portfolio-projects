# Flaky Pipeline Gate

## Overview
Flaky Pipeline Gate evaluates job history and determines whether a failing test is likely flaky or a real regression.

## Key Feature
Its unique feature is a merge decision that discounts failures likely caused by flaky tests rather than code changes.

## Architecture
- `src/main.py`: Flakiness scoring and gate decision logic.
- `tests/test_smoke.py`: Covers score computation and gate behavior.
- `Makefile`: Run/test shortcuts.

## Use Cases
- Prevent blocked merges from historically flaky tests.
- Alert teams when previously stable tests become unreliable.
- Standardize CI gating policy across services.

## Usage
```bash
make run
make test
```

## Control Flow
1. Ingest recent pass/fail outcomes.
2. Compute rolling flakiness score.
3. Compare score against threshold.
4. Return `block` or `allow` decision for merge automation.

## Project Structure
```text
flaky-pipeline-gate/
  README.md
  Makefile
  src/
    main.py
  tests/
    test_smoke.py
```
