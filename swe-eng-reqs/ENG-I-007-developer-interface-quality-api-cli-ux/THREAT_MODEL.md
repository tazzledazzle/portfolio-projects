# Threat / failure notes — ENG-I-007

- Trust boundary: untrusted CLI argv (T-5-04). Errors must be clear without
  stack traces or secret leakage; never echo credentials from the environment.
- Authn/z: demo binds locally; production would require OIDC/RBAC.
- Multi-tenant isolation: enforce quotas before exposing shared endpoints.
- Supply chain: pin base images; sign artifacts in registry slices.
- Failure modes: dependency outage should degrade gracefully (see `/v1/demo` proof fields).
- Boundary: this folder does not implement OpenAPI or rate-limit middleware.
