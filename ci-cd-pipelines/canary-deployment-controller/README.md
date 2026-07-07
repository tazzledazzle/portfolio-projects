# Canary Deployment Controller

## Overview
Canary Deployment Controller models progressive traffic shifting and automated rollback decisions using canary SLO checks.

## Key Feature
The unique capability is automatic rollback based on error rate and p99 latency thresholds during progressive rollout.

## Architecture
- `src/main.py`: Progressive shift + SLO evaluation logic.
- `tests/test_smoke.py`: Validates promotion and rollback outcomes.
- `Makefile`: Run and test commands.

## Use Cases
- Safely roll out high-risk releases.
- Enforce SLO-aware deployment policies.
- Automate rollback decisions consistently.

## Usage
```bash
make run
make test
```

## Control Flow
1. Shift traffic to canary by rollout step.
2. Pull current error rate and p99 values.
3. Compare against configured thresholds.
4. Continue rollout or trigger rollback.

## Project Structure
```text
canary-deployment-controller/
  README.md
  Makefile
  src/
    main.py
  tests/
    test_smoke.py
```
