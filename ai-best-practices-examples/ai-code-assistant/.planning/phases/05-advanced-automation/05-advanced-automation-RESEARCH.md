# Phase 5: Advanced Automation - Research

**Researched:** 2026-03-31
**Domain:** Checkpointed autonomous CLI workflows (plan/execute/verify), optional parallelism, rollback safety
**Confidence:** HIGH

## User Constraints

No `*-CONTEXT.md` file found for this phase. Constraints inferred from roadmap:
- Implement `plan -> execute -> verify` loops with checkpoints.
- Support optional parallel task execution only for independent operations.
- Enforce rollback-first behavior for high-risk changes.

## Summary

Phase 5 should be implemented as a deterministic orchestration layer on top of the existing CLI core, not as a fully autonomous free-form agent loop. The practical path is a typed execution graph with explicit checkpoints and persisted run state (`run_id`, task states, checkpoint records, verification results). Each loop iteration should be: load plan state -> execute next task(s) -> run scoped verification -> gate continuation based on policy.

Parallelism should be opt-in and constrained to independent tasks in the same stage. Python's `concurrent.futures.ThreadPoolExecutor` provides sufficient and stable primitives for this CLI use case (bounded workers, futures, cancellation/shutdown handling) while keeping complexity lower than distributed orchestration frameworks.

Rollback-first safety should treat high-risk actions as transactional from the orchestrator point of view: create a pre-change checkpoint (git ref + snapshot metadata), perform changes in isolated worktree/branch when possible, verify, then either promote or rollback automatically. This aligns with Git safeguards and keeps failure recovery predictable.

**Primary recommendation:** Build a small `run_orchestrator` module using `ThreadPoolExecutor` + git-backed checkpoints before introducing any multi-agent runtime.

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Python stdlib `concurrent.futures` | Python 3.10+ (project baseline) | Optional bounded parallel task execution | Official stdlib API, stable executor/future model |
| Python stdlib `subprocess` | Python 3.10+ | Verification command execution (`pytest`, lint, type-check) with timeout/check handling | Robust process control with explicit failure semantics |
| Git CLI (`git worktree`, `git diff`, `git rev-parse`) | Git 2.x | Checkpointing and rollback primitives | Battle-tested rollback/recovery model and safety guards |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `pytest` | `>=8.0.0` (already in project) | Orchestrator unit/integration/e2e verification | Required for all phase acceptance tests |
| `pytest-xdist` | latest | Validate test suite parallel-safety and speed in CI only | Optional for CI acceleration, not required by runtime |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| stdlib executors | `asyncio.TaskGroup` | Strong structured concurrency, but fail-fast cancellation behavior is often too aggressive for mixed independent tasks |
| git CLI subprocess | `GitPython` | Cleaner API but extra dependency surface; CLI is already available and auditable |
| local checkpoint files | SQLite run-state store | Better querying but unnecessary complexity for initial phase |

**Installation:**
```bash
pip install -e ".[dev]"
pip install pytest-xdist
```

## Architecture Patterns

### Recommended Project Structure
```text
src/ai_code_assistant/
├── orchestration/            # run loop, scheduler, checkpoint manager
├── verification/             # verification pipeline runners
├── safety/                   # risk scoring, rollback policy, recovery
└── cli.py                    # command routing and options
```

### Pattern 1: Checkpointed Run State Machine
**What:** Explicit task lifecycle (`planned`, `running`, `verified`, `failed`, `rolled_back`) persisted per run.
**When to use:** All automated multi-step operations.
**Example:**
```python
# Source: project design pattern (derived), subprocess semantics from Python docs
for stage in plan.stages:
    checkpoint = checkpoint_manager.create(stage=stage.name)
    result = executor.run_stage(stage, parallel=cfg.parallel)
    verify = verifier.run(stage.verify_commands, timeout=cfg.verify_timeout_s)
    if not verify.ok:
        safety.rollback(checkpoint)
        mark_stage_failed(stage, reason=verify.error)
        break
```

### Pattern 2: Bounded Optional Parallel Stage Execution
**What:** Execute only dependency-free tasks concurrently with fixed worker cap.
**When to use:** Stage has 2+ tasks with no shared output files/commands.
**Example:**
```python
# Source: https://docs.python.org/3/library/concurrent.futures.html
with ThreadPoolExecutor(max_workers=config.max_parallel_tasks) as pool:
    futures = [pool.submit(run_task, task) for task in ready_tasks]
    for fut in as_completed(futures):
        task_result = fut.result()
        collect(task_result)
```

### Pattern 3: Rollback-First High-Risk Guard
**What:** Create rollback point before mutating operations above risk threshold.
**When to use:** Multi-file writes, refactors, shell commands with destructive potential.
**Example:**
```python
# Source: https://git-scm.com/docs/git-worktree and Python subprocess docs
pre = git.capture_head()
safety.create_checkpoint_ref(pre)
try:
    run_high_risk_change()
    verification.require_pass()
except Exception:
    git.restore_to(pre)
    raise
```

### Anti-Patterns to Avoid
- **Implicit mutable state:** Do not infer progress from filesystem alone; always persist run state explicitly.
- **Unbounded parallelism:** Do not map all tasks to default worker count; cap workers and gate by dependency graph.
- **Verify-last design:** Do not batch all verification at the end; verify per checkpoint.
- **No pre-change snapshot:** Never run high-risk mutations without a rollback point.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Concurrency primitives | Custom thread lifecycle manager | `ThreadPoolExecutor` | Handles worker lifecycle and shutdown semantics reliably |
| Process execution | Ad-hoc `os.system` wrappers | `subprocess.run(..., check=True, timeout=...)` | Structured errors and timeout safety |
| Repository safety lifecycle | Custom VCS state store | Git refs/worktrees + audit log | Native safety checks and recovery primitives |

**Key insight:** Reliability here comes from orchestration policy, not novel runtime machinery.

## Common Pitfalls

### Pitfall 1: Parallel Tasks Touching Same Files
**What goes wrong:** Nondeterministic edits, flaky verification, hidden races.
**Why it happens:** Dependency graph not enforced before scheduling.
**How to avoid:** Build file-level conflict detection and schedule only non-overlapping write sets.
**Warning signs:** Same file modified by multiple tasks in one stage, intermittent failures.

### Pitfall 2: Verification Commands Hanging
**What goes wrong:** Run never completes; checkpoint becomes stale.
**Why it happens:** Missing process timeout and kill handling.
**How to avoid:** Always set per-command timeout and mark verification timeout as hard failure.
**Warning signs:** Stage execution exceeds configured SLA with no output progress.

### Pitfall 3: Partial Rollback
**What goes wrong:** HEAD restored but generated files/side effects remain.
**Why it happens:** Rollback only tracks git state, not temp artifacts.
**How to avoid:** Track artifact paths in checkpoint metadata and clean during rollback.
**Warning signs:** Dirty tree remains after rollback, non-git files left behind.

## Code Examples

Verified patterns from official sources:

### Strict subprocess verification
```python
# Source: https://docs.python.org/3/library/subprocess.html
completed = subprocess.run(
    ["pytest", "-q"],
    check=True,
    timeout=300,
    capture_output=True,
    text=True,
)
```

### Executor lifecycle with context manager
```python
# Source: https://docs.python.org/3/library/concurrent.futures.html
with ThreadPoolExecutor(max_workers=4) as executor:
    futures = [executor.submit(run_task, t) for t in tasks]
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Single-pass automation (plan once, run all, verify once) | Iterative checkpointed loops | Ongoing best practice (2024-2026 tooling) | Faster fault isolation, safer autonomy |
| Monolithic sequential workers | Bounded optional parallel stages | Mature stdlib/executor guidance | Better throughput without losing determinism |
| Best-effort undo scripts | Git-backed rollback-first checkpoints | Longstanding Git operational practice | Predictable recovery for risky changes |

**Deprecated/outdated:**
- Long-running uncontrolled thread pools for CLI jobs: stdlib docs caution against thread pools for long-running tasks without explicit lifecycle management.

## Required Tests

1. **Run loop checkpoint progression (unit)**
   - Given staged plan with 3 tasks, assert state transitions and checkpoint records after each stage.
2. **Verification gate failure triggers rollback (integration)**
   - Force verifier failure; assert repository returns to pre-change commit/tree and run marked failed/rolled_back.
3. **Parallel execution for independent tasks (unit/integration)**
   - Two independent tasks complete in parallel; assert both results recorded and no ordering dependency.
4. **Conflict detection disables parallelization (unit)**
   - Two tasks target same file; assert scheduler switches to serial or rejects plan with policy error.
5. **Timeout handling in verification commands (integration)**
   - Simulate hanging subprocess; assert timeout, task failure, rollback trigger, and audit event.
6. **Checkpoint resume behavior (integration)**
   - Interrupt run mid-stage; rerun with same `run_id`; assert resume from last stable checkpoint.
7. **High-risk policy enforcement (unit)**
   - Risk score above threshold requires checkpoint creation before execute; assert no mutation if checkpoint creation fails.
8. **Audit completeness (unit)**
   - Assert audit log contains: run_started, checkpoint_created, task_completed/failed, verify_passed/failed, rollback_applied.

## Open Questions

1. **Checkpoint persistence format**
   - What we know: JSON file is enough initially.
   - What's unclear: expected run volume and retention strategy.
   - Recommendation: start with JSONL per run, add SQLite only if query/reporting needs emerge.
2. **Risk scoring thresholds**
   - What we know: high-risk actions must be rollback-first.
   - What's unclear: exact scoring rubric and user override policy.
   - Recommendation: define fixed v1 rubric in policy config before coding.

## Sources

### Primary (HIGH confidence)
- Python docs: `asyncio` tasks and TaskGroup semantics - https://docs.python.org/3/library/asyncio-task.html
- Python docs: `concurrent.futures` executor APIs - https://docs.python.org/3/library/concurrent.futures.html
- Python docs: `subprocess` timeout/check behavior - https://docs.python.org/3/library/subprocess.html
- Git official docs: `git worktree` safeguards - https://git-scm.com/docs/git-worktree

### Secondary (MEDIUM confidence)
- pytest-xdist official docs (parallel test execution, limitations, changelog) - https://pytest-xdist.readthedocs.io/en/stable/

### Tertiary (LOW confidence)
- None.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - based on official Python and Git documentation.
- Architecture: MEDIUM-HIGH - patterns are prescriptive synthesis grounded in official primitives.
- Pitfalls: MEDIUM - based on common failure modes plus documented cancellation/timeout behavior.

**Research date:** 2026-03-31
**Valid until:** 2026-04-30
