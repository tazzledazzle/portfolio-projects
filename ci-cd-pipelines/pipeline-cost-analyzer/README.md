# Pipeline Cost Analyzer

## Overview
Pipeline Cost Analyzer attributes CI runtime cost to workflows, services, and teams, then suggests optimization opportunities.

## Key Feature
The unique capability is team/service/job-level cost attribution with optimization hints based on runner usage patterns.

## Architecture
- `src/main.py`: Cost aggregation and recommendation logic.
- `tests/test_smoke.py`: Checks attribution totals and optimization output.
- `Makefile`: Run/test commands.

## Use Cases
- Identify most expensive pipeline jobs.
- Detect runner over-provisioning.
- Prioritize CI optimization backlog by spend impact.

## Usage
```bash
make run
make test
```

## Control Flow
1. Ingest minutes by runner type and workflow.
2. Calculate total cost per owner dimension.
3. Rank high-cost jobs.
4. Emit targeted optimization recommendations.

## Project Structure
```text
pipeline-cost-analyzer/
  README.md
  Makefile
  src/
    main.py
  tests/
    test_smoke.py
```
