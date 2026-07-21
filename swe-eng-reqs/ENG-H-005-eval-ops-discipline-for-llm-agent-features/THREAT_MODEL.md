# Threat / failure notes — ENG-H-005

- **T-6-05 — prompt-injection fixture tampering:** fixture text is treated only
  as untrusted scoring data and is never executed as an instruction. The
  known-bad injection output must fail and increment
  `failure_fixtures_caught`.
- Online-sim accepts only an in-process callable selected by the service; it
  does not accept provider URLs, execute fixture text, or perform network I/O.
- Invalid JSON and unknown eval modes are rejected before running the harness.
- Authn/z: demo binds locally; production would require OIDC/RBAC.
- Multi-tenant isolation: enforce quotas before exposing shared endpoints.
- Supply chain: runtime uses Python stdlib only; no model-provider, agent, or
  MCP SDK is installed.
- Failure modes: dependency outage should degrade gracefully (see `/v1/demo` proof fields).
