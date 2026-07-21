# Threat / failure notes — ENG-E-017

- **T-6-03 — unauthorized tool call / elevation:** `tools/call` requires a
  token-inspired `tools:read` scope before dispatch. The registry rejects every
  mutating tool; the approval policy gateway belongs to ENG-N-015.
- **T-6-04 — audit gaps and disclosure:** every allowed or denied call appends
  an audit event. Keys named `secret`, `token`, `password`, or `api_key` are
  recursively redacted before storage.
- Invalid methods and parameters return bounded JSON-RPC errors; unknown tools
  are denied and audited.
- Authn/z: the demo token is a deterministic simulator; production would
  require an external identity provider, token verification, and RBAC.
- Multi-tenant isolation: enforce quotas before exposing shared endpoints.
- Supply chain: runtime uses Python stdlib only; the official MCP SDK and live
  model-provider clients are intentionally absent.
- Failure modes: dependency outage should degrade gracefully (see `/v1/demo` proof fields).
