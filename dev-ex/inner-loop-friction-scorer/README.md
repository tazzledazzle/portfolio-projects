# Inner Loop Friction Scorer

## Overview
A scoring engine that combines CI telemetry, Git lifecycle signals, and self-reported metrics into a single composite inner-loop friction score per engineering team.

## Key Feature
A transparent weighted friction taxonomy that converts multiple bottlenecks (setup, build, test, review, deploy) into one comparable team score.

## Architecture
- Python scoring service for deterministic composite calculations
- Adapters for CI events, Git metadata, and self-reported inputs
- Core weighting model to produce normalized friction scores
- Output adapters for trend stores and downstream dashboards
- Configurable weighting profiles for organization-specific priorities

## Use Cases
- Identify which teams experience highest development friction
- Measure impact of build cache or test parallelization investments
- Detect regressions in PR review latency over time
- Prioritize platform backlog using quantified friction outcomes

## Usage
```bash
make install
make run
```

Run tests:
```bash
make test
```

## Control Flow
1. Telemetry adapters collect setup, build, test, review, and deploy signals.
2. Inputs are normalized into a shared metrics contract.
3. Composite engine applies category weights and computes friction scores.
4. Results are emitted per team and period for trend analysis.
5. Platform teams inspect high-friction categories and execute remediations.

## Project Structure
- `src/friction_scorer/main.py`: CLI/service entrypoint
- `src/friction_scorer/adapters/`: Data source adapters
- `src/friction_scorer/core/`: Composite scoring logic
- `tests/`: Scoring and execution tests
- `scripts/`: Local execution scripts
- `docs/`: Metrics model documentation
