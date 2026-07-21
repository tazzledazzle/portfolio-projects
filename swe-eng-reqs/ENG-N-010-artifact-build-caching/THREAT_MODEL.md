# Threat / failure notes — ENG-N-010

- Authn/z: demo binds locally; production would require OIDC/RBAC.
- Multi-tenant isolation: enforce quotas before exposing shared endpoints.
- Supply chain: pin base images; sign artifacts in registry slices.
- Failure modes: dependency outage should degrade gracefully (see `/v1/demo` proof fields).
