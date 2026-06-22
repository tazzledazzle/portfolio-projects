# Dead Code Surface Reporter

## Overview
This project identifies cleanup candidates by combining unreachable code detection with recency scoring from git history.

## Architecture
- `src/analysis`: symbol and call-graph extraction.
- `src/git`: blame recency enrichment.
- `src/scoring`: weighted prioritization logic.
- `src/reporting`: cleanup backlog output.
- `config`: score weights and age thresholds.

## Use Cases
- Generate quarterly dead-code cleanup backlogs.
- Prioritize removals with low risk and low activity.
- Provide measurable debt reduction targets.

## Usage
1. Run analysis against target module.
2. Join call-graph results with git recency metadata.
3. Emit prioritized JSON/Markdown reports.

## Control Flow
1. Build symbol graph and mark reachable nodes.
2. Select unreachable symbols.
3. Fetch last-touch dates with git blame.
4. Score symbols by staleness + reachability confidence.
5. Publish ranked backlog.

## Project Structure
```text
dead-code-surface-reporter/
  .github/workflows/ci.yml
  config/scoring.yml
  scripts/run-report.sh
  src/analysis/call_graph.py
  src/git/blame_recency.py
  src/scoring/prioritizer.py
  src/reporting/report_writer.py
  tests/test_prioritizer.py
```
