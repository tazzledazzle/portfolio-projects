# Feature Landscape

**Domain:** AI code assistant CLI (Python-first)
**Researched:** 2026-03-31

## Table Stakes

Features users expect. Missing = product feels incomplete.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Interactive + non-interactive mode (`ask`, `exec`) | All major CLIs support chat + script automation | Med | Must support CI and local workflows |
| Multi-file edit with diff preview | Core value prop vs plain chat | High | Needs patch safety and rollback |
| Git-aware workflows (branch/checkpoint/commit assist) | Users want reversible AI edits | Med | Aider and commercial tools normalize this |
| Approval policies + sandbox modes | Security baseline in modern agentic CLIs | High | Read-only, workspace-write, full-access |
| Provider flexibility (cloud APIs + local models) | Cost/privacy/perf tradeoffs are frequent | Med | Different teams have different constraints |
| Repo context memory files (`AGENTS.md`/project instructions) | Persistent guidance is now standard | Low | Crucial for repeatability |

## Differentiators

Features that set product apart. Not expected, but valued.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Python-aware semantic refactor mode | Safer large edits in Python-heavy repos | High | Leverage AST + tests + type checks |
| Security-first execution profile | Makes agent usable in regulated orgs | High | Built-in secret redaction, command risk scoring |
| Deterministic "plan then execute" with checkpoints | Better trust and auditability | Med | Reduces "AI did mysterious stuff" failures |
| Cost/performance-aware model router | Cuts spend while preserving quality | Med | Route simple tasks to cheaper models |

## Anti-Features

Features to explicitly NOT build.

| Anti-Feature | Why Avoid | What to Do Instead |
|--------------|-----------|-------------------|
| "Fully autonomous YOLO by default" | High risk of destructive actions | Default to on-request approvals + explicit escalation |
| Massive plugin marketplace in MVP | Supply-chain and quality overhead too early | Curated built-in tools + signed plugin allowlist later |
| Cloud-only memory by default | Privacy/compliance blocker for many CLI users | Local-first state, optional encrypted sync |
| "One prompt creates whole app" marketing mode | High hallucination and low reliability for real teams | Task-scoped workflows: fix bug, write tests, refactor module |

## Feature Dependencies

```text
Sandbox + Approval Policy -> Safe Command Execution -> Autonomous Multi-step Tasks
Git Safety + Diff Engine -> Multi-file Edits -> Refactor/Modernization Workflows
Model Adapter Layer -> Cost Router -> Team Policy Enforcement
```

## MVP Recommendation

Prioritize:
1. Interactive + scripted modes with safe defaults
2. Git-aware multi-file edits with reviewable patch output
3. Security baseline (approval gates, secrets redaction, local audit log)

Defer: Marketplace-style third-party plugin ecosystem: strong value later, high early risk.

## Sources

- [OpenAI Codex CLI features/security](https://developers.openai.com/codex/cli)
- [Claude Code overview and workflows](https://code.claude.com/docs/en/overview)
- [Gemini CLI capabilities](https://google-gemini.github.io/gemini-cli/)
- [Aider capabilities](https://aider.chat/docs/)
- [GitHub Copilot CLI docs](https://docs.github.com/copilot/using-github-copilot/using-github-copilot-in-the-command-line)
