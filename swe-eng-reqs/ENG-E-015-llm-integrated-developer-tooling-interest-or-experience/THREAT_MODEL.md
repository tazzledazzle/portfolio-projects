# Threat / failure notes — ENG-E-015

| Threat | Mitigation |
|---|---|
| Secret leakage from failure payloads | Secret-shaped fields (`api_key`, `authorization`, `password`, `secret`, `token`) are redacted before fixture completion and never copied to audit records. |
| Accidental live-provider use | `OfflineFixtureLLM` has `live=false`; the slice has no provider SDK, network client, or API-key lookup. |
| Oversized or malformed summarize input | HTTP requests are limited to 64 KiB and validated as non-empty failure objects. |
| Unauthenticated public use | The demo binds locally; a production deployment would add OIDC/RBAC and tenant quotas. |
| Supply-chain substitution | Runtime is Python stdlib; `requirements.txt` contains pytest only. |

Failure-log text is treated as untrusted content and can only become summary input. This slice
does not own or invoke tools, agent loops, workflow actions, or mutations.
