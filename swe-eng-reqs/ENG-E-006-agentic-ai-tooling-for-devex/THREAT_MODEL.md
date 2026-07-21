# Threat / failure notes — ENG-E-006

| Threat | Mitigation |
|---|---|
| T-6-06 Auto-execute mutate | Propose-only path; `executed=false` always; mutating actions set `requires_approval=true`; no approval grant store (N-015 owns grants). |
| Untrusted pipeline ids | Unknown ids raise `KeyError` / HTTP 404; fixtures validated before diagnosis. |
| Accidental live-provider use | No provider SDK, network client, or API-key lookup; honesty `live_provider=false`. |
| Oversized diagnose body | HTTP requests limited to 64 KiB and must be JSON objects. |
| Unauthenticated public use | Demo binds locally; production would add OIDC/RBAC and tenant quotas. |
| Supply-chain substitution | Runtime is Python stdlib; `requirements.txt` contains pytest only. |

This slice diagnoses and proposes. It does not execute mutations, host an MCP
server, run an agent Plan-Execute loop, or implement a full policy gateway.
