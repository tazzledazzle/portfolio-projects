# Architecture Patterns

**Domain:** AI code assistant CLI
**Researched:** 2026-03-31

## Recommended Architecture

Python-first, local-first agent runtime with explicit trust boundaries:

`CLI Shell -> Planner -> Tool Executor (sandboxed) -> Verifier (tests/lint/security) -> Git Patch Layer -> User Approval`

### Component Boundaries

| Component | Responsibility | Communicates With |
|-----------|---------------|-------------------|
| CLI Frontend | Commands, flags, prompt UX, streaming output | Session Manager, Planner |
| Session Manager | Conversation state, checkpoints, history | Planner, Policy Engine, Storage |
| Planner | Task decomposition and step plans | Tool Executor, Verifier |
| Policy Engine | Approval rules, allowed tools, risk scoring | Tool Executor, UI |
| Tool Executor | File edits, shell commands, web/tool calls in sandbox | Sandbox Adapter, Verifier |
| Verifier | Runs tests/lint/security checks and evaluates deltas | Git Layer, UI |
| Git Layer | Diff generation, branch safety, rollback/checkpoints | CLI Frontend |
| Provider Adapter | Model abstraction (cloud/local) | Planner, Session Manager |

### Data Flow

1. User invokes command or interactive task.
2. Planner proposes structured steps and required permissions.
3. Policy engine classifies actions (safe, review-required, blocked).
4. Executor runs allowed steps in sandboxed scope.
5. Verifier runs targeted checks (`tests`, `lint`, `typecheck`, optional `semgrep`).
6. Git layer presents patch + confidence + check results.
7. User accepts/rejects, then optional commit PR handoff.

## Patterns to Follow

### Pattern 1: Capability-scoped execution
**What:** Every tool call runs with minimal needed privileges.
**When:** Always; especially for shell/network/MCP-style integrations.
**Example:**
```python
if action.requires_network and not policy.network_enabled:
    raise PermissionError("network disabled by policy")
```

### Pattern 2: Verify-after-edit loop
**What:** Treat generated edits as candidate patches until verified.
**When:** Any code-modifying task.
**Example:**
```python
apply_patch(patch)
results = run_checks(["pytest -q", "ruff check .", "mypy ."])
if not results.ok:
    rollback_patch()
```

## Anti-Patterns to Avoid

### Anti-Pattern 1: Unbounded shell autonomy
**What:** Allowing arbitrary command execution by default.
**Why bad:** Escalates prompt-injection and destructive-command risk.
**Instead:** Explicit approval gates + denylist + sandbox profile.

### Anti-Pattern 2: Stateless "single-shot" edits
**What:** No checkpointing or rollback.
**Why bad:** Hard to recover from almost-correct AI changes.
**Instead:** Mandatory checkpoints before risky operations.

## Scalability Considerations

| Concern | At 100 users | At 10K users | At 1M users |
|---------|--------------|--------------|-------------|
| Model cost | Per-user key acceptable | Introduce routing and caching | Strict policy routing + budget enforcement |
| Security governance | Local config enough | Team-managed policy bundles | Centralized org policy + audit pipeline |
| Plugin/tool trust | Curated built-ins | Signed extension allowlist | Full trust framework + continuous vetting |
| Supportability | Manual logs | Structured telemetry optional | Compliance-grade telemetry + redaction controls |

## Sources

- [Codex approvals/sandbox architecture](https://developers.openai.com/codex/agent-approvals-security)
- [Gemini CLI sandboxing and trusted workflow docs](https://google-gemini.github.io/gemini-cli/docs/cli/sandbox.html)
- [Claude Code tooling workflow overview](https://code.claude.com/docs/en/overview)
