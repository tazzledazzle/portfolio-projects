# AI Code Assistant Roadmap

## Product Goal

Build a secure, Python-first CLI coding assistant that can generate tests, apply safe changes, and progressively automate development workflows with strong policy controls.

## Guiding Principles

- Local-first by default; explicit opt-in for external integrations.
- Verify before write whenever possible.
- Keep human approval in the loop for risky operations.
- Optimize for reliability and reproducibility over raw autonomy.

## Phase Plan

### Phase 1: Secure CLI Foundation (Now)

Outcome: establish trust and safe defaults.

1. Add execution profiles (`read-only`, `workspace-write`, `full-access`).
2. Enforce profile restrictions in file mutation flows.
3. Add JSON output mode for CI-friendly runs.
4. Add operation audit logging (local file).

Success metrics:

- 0 unauthorized writes in `read-only` profile.
- 100% profile behavior covered by tests.
- >= 95% command success in local smoke runs.

### Phase 2: Python Workflow Depth

Outcome: increase quality of generated edits.

1. AST-aware source analysis for better test generation.
2. Structured test pyramid generation (`unit`, `integration`, `e2e` templates).
3. Deterministic "fix failing generated test" loop.

Success metrics:

- >= 70% generated-test pass rate on sample repos.
- <= 10% regression rate for AST-based rewrites.

### Phase 3: Policy and Trust Controls

Outcome: support team and enterprise usage.

1. Configurable policy file (`assistant-policy.toml`).
2. Risk-scored shell/tool actions with approval gates.
3. Secret redaction in prompts/logs.

Success metrics:

- 0 high-risk actions without approval.
- 0 known secret leakage incidents in test harness.

### Phase 4: Integrations and Extension Surface

Outcome: connect to real delivery workflows.

1. GitHub/PR metadata ingestion.
2. Signed extension manifest and capability scopes.
3. Headless pipeline mode for CI.

Success metrics:

- >= 95% extension install success (signed).
- < 3% extension-induced runtime failures.

### Phase 5: Advanced Automation

Outcome: checkpointed multi-step autonomous execution.

1. Plan -> execute -> verify loops with checkpoints.
2. Optional parallel task agents for independent operations.
3. Rollback-first safety for high-risk changes.

Success metrics:

- >= 60% fully automated task completion on benchmark flows.
- >= 25% reduction in time-to-merge against baseline.

## Prioritized Backlog

1. Execution profiles and enforcement.
2. Git-safe patch preview and rollback.
3. Verification pipeline (tests/lint/type-check).
4. JSON mode output for automation.
5. AST-aware Python generation quality.
6. Secret/context guardrails.
7. Policy file support.
8. Multi-model routing.
9. Risk-scored command execution.
10. Signed extension framework.

## Implementation Start (Step-by-Step)

Step 1 (current): execution profiles + enforcement in CLI write paths.
Step 2: JSON output mode for deterministic CI integration.
Step 3: audit logging for all write and generation actions.
