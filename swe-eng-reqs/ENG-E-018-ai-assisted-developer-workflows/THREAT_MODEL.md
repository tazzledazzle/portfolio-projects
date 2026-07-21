# Threat / failure notes — ENG-E-018

| Threat | Mitigation |
|---|---|
| T-6-07 Skip-approval | Stage gate blocks `approved` without `approve()`; `advance()` raises at `awaiting_approval`; invalid early approve rejected. |
| Illegal stage jumps | FSM only advances adjacent stages; unknown workflow ids raise `KeyError` / HTTP 404. |
| Accidental live-provider use | No provider SDK or API-key lookup; honesty `live_provider=false`. |
| Oversized create body | HTTP requests limited to 64 KiB and must be JSON objects. |
| Unauthenticated public use | Demo binds locally; production would add OIDC/RBAC and tenant quotas. |
| Supply-chain substitution | Runtime is Python stdlib; `requirements.txt` contains pytest only. |

This slice owns the multi-stage workflow with an approval gate. It does not
implement OfflineFixtureLLM summarization, ToolRegistry pedagogy, MCP auth,
or a full mutating policy grant store (N-015).
