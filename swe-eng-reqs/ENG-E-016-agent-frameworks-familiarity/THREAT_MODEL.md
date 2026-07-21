# Threat / failure notes — ENG-E-016

| Threat | Mitigation |
|---|---|
| Mutating tool runs without human approval | `ToolRegistry.call` rejects every mutating tool unless `approved=True`, before invoking its handler. |
| Unbounded agent execution | `run_agent` enforces a hard maximum of five Plan-Execute steps. |
| Unknown or malformed tool calls | The registry denies unknown names and arguments missing schema fields. |
| Oversized or malformed agent input | HTTP requests are limited to 64 KiB and validate the goal and step limit. |
| Unauthenticated public use | The demo binds locally; a production deployment would add OIDC/RBAC and tenant quotas. |
| Agent-framework supply-chain substitution | Runtime is Python stdlib; `requirements.txt` contains pytest only. |

The mutating approval flag demonstrates a local safety boundary only. Full approval grants and
policy-gateway behavior belong to ENG-N-015 and are intentionally absent here.
