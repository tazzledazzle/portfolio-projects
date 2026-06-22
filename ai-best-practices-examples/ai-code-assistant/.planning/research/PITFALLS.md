# Domain Pitfalls

**Domain:** AI code assistant CLI
**Researched:** 2026-03-31

## Critical Pitfalls

Mistakes that cause rewrites or major issues.

### Pitfall 1: Prompt injection through fetched content/tools
**What goes wrong:** Assistant follows malicious instructions hidden in docs/web pages/repo text.
**Why it happens:** No trust boundary between "data" and "instructions."
**Consequences:** Unauthorized commands, data exfiltration, policy bypass.
**Prevention:** Treat external content as untrusted, isolate tool outputs, require explicit approvals for side effects.
**Detection:** Unexpected command suggestions, sudden privilege escalation requests.

### Pitfall 2: Secret leakage in prompts, logs, and telemetry
**What goes wrong:** API keys/tokens/private code leak to model providers or log backends.
**Why it happens:** Raw environment/context passed without redaction.
**Consequences:** Credential compromise and compliance violations.
**Prevention:** Redact known secret patterns before send/log, disable prompt logging by default, local encrypted history.
**Detection:** DLP scans on logs, canary tokens, abnormal token usage.

### Pitfall 3: Supply-chain risk via plugins/MCP/local servers
**What goes wrong:** Malicious extension or startup command executes code locally.
**Why it happens:** One-click install/trust with poor verification.
**Consequences:** Host compromise, data theft, persistent backdoors.
**Prevention:** Signed plugins, explicit command preview, allowlist registries, least privilege scopes.
**Detection:** Integrity checks, provenance verification (SLSA-style), unexpected outbound traffic.

## Moderate Pitfalls

### Pitfall 1: "Almost-right code" quality debt
**What goes wrong:** AI outputs pass quick review but create long-term regressions.
**Prevention:** Verify-after-edit loop with tests, lint, type-check, and policy checks.

### Pitfall 2: Overly broad default permissions
**What goes wrong:** Convenience defaults become unsafe org-wide norms.
**Prevention:** Secure-by-default profiles (`read-only` or workspace-write + on-request approvals).

## Minor Pitfalls

### Pitfall 1: Poor latency UX in large repos
**What goes wrong:** Users abandon tool due to slow planning/context loading.
**Prevention:** Incremental indexing, scoped context windows, cached repo map.

## Phase-Specific Warnings

| Phase Topic | Likely Pitfall | Mitigation |
|-------------|---------------|------------|
| MVP automation | Too much autonomy too early | Require manual approval on all write + shell actions |
| Plugin ecosystem | Trusting unsigned third-party tools | Signed manifests, disabled by default |
| Team rollout | Missing auditability | Local event log + optional OTel export with redaction |
| Cloud features | Data residency/compliance drift | Local-first mode and explicit regional routing |

## Sources

- [OWASP LLM Top 10 / GenAI security project](https://genai.owasp.org/llm-top-10/)
- [MCP security best practices](https://modelcontextprotocol.io/specification/2025-11-25/basic/security_best_practices)
- [Codex security and approval guidance](https://developers.openai.com/codex/agent-approvals-security)
- [SLSA supply-chain framework](https://slsa.dev/)
