# Technology Stack

**Project:** ai-code-assistant
**Researched:** 2026-03-31

## Recommended Stack

### Core Framework
| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| Python + Typer | Python 3.11+, Typer latest | CLI UX, command routing, subcommands | Python-first fit, fast implementation, strong ecosystem, easy distribution via pipx/Homebrew wrappers |
| prompt_toolkit + Rich/Textual-lite UI patterns | latest | Interactive terminal UX, streaming, diff views | Matches user expectations set by Codex/Claude/Gemini terminal apps without overbuilding full TUI at MVP |

### Database
| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| SQLite (default local state) | latest | Session memory, approvals, audit trail, cached summaries | Offline-first baseline, zero infra, deterministic local behavior |
| Postgres (optional team mode) | latest | Shared sessions, org policy state, multi-user telemetry | Needed only beyond solo developer mode |

### Infrastructure
| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| Local subprocess sandboxing + policy engine | latest | Safe command execution | Codex/Claude/Gemini all emphasize sandbox + approval gates as table stakes for agentic CLIs |
| Provider adapters (OpenAI/Anthropic/Google/local model gateway) | latest | Model routing and fallback | Market is multi-provider and price/perf changes quickly |

### Supporting Libraries
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| tree-sitter bindings | latest | AST-aware code edits | Required once you move beyond single-file regex edits |
| gitpython or direct git subprocess wrappers | latest | Atomic patching, branch safety, rollback | Required for any non-trivial automated code change |
| pydantic | latest | Config/schema validation | Needed for safe policy/config loading and plugin manifests |
| semgrep (optional integration) | latest | Security scan pass after edits | Add in phase 2+ for guardrails before commit |

## Alternatives Considered

| Category | Recommended | Alternative | Why Not |
|----------|-------------|-------------|---------|
| Runtime | Python | Rust | Rust is great for performance, but Python wins for speed of iteration and ecosystem for AST/security/dev tooling |
| UI | Typer + lightweight interactive loop | Full custom TUI framework from day 1 | High complexity early; delays shipping core reliability |
| State | SQLite local-first | Cloud-only backend | Increases privacy/compliance risk and onboarding friction for CLI-first users |
| Model strategy | Multi-provider adapter | Single vendor lock-in | Competitors increasingly support provider choice; lock-in is a business and reliability risk |

## Installation

```bash
# Core
pip install typer rich prompt-toolkit pydantic

# Optional advanced integrations
pip install tree-sitter gitpython
```

## Sources

- [OpenAI Codex CLI docs](https://developers.openai.com/codex/cli) (official)
- [OpenAI Codex agent approvals & security](https://developers.openai.com/codex/agent-approvals-security) (official)
- [Claude Code overview](https://code.claude.com/docs/en/overview) (official)
- [Gemini CLI docs](https://google-gemini.github.io/gemini-cli/) (official)
- [Gemini CLI sandboxing](https://google-gemini.github.io/gemini-cli/docs/cli/sandbox.html) (official)
- [Aider docs/home](https://aider.chat/) (official)
- [GitHub Copilot CLI install/docs](https://docs.github.com/en/copilot/how-tos/set-up/install-copilot-in-the-cli) (official)
