# Threat / failure notes — ENG-N-015

- Authn/z: demo binds locally; production would require OIDC/RBAC.
- Multi-tenant isolation: enforce quotas before exposing shared endpoints.
- Supply chain: pin base images; sign artifacts in registry slices.
- Failure modes: dependency outage should degrade gracefully (see `/v1/demo` proof fields).

| ID | Threat | Mitigation |
|---|---|---|
| T-6-08 | Mutating `tools/call` without human authorization | Deny by default and require an opaque grant bound to the exact intent digest. |
| T-6-09 | Prompt-injection text in arguments attempts to elevate privilege | Arguments never act as approval; only the separately issued digest-bound token is accepted. |
| T-6-09A | Secrets leak through append-only audit records | Recursively redact secret, token, password, and api_key fields before append. |

The policy gateway is a local portfolio simulator. Production deployment would
also authenticate the approving human, expire and consume grants, and persist
tamper-evident audit records.
