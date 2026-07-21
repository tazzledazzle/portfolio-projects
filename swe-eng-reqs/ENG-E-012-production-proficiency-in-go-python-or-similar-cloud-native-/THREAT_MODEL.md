# Threat / failure notes — ENG-E-012

- Spoofing of OR claims (T-5-05): Info/demo must expose `or_semantics=true` and
  `language=go` explicitly; do not claim paired Go+Python as mandatory.
- Authn/z: demo binds locally; production would require OIDC/RBAC.
- Multi-tenant isolation: enforce quotas before exposing shared endpoints.
- Supply chain: pin base images; sign artifacts in registry slices (stdlib only here).
- Failure modes: dependency outage should degrade gracefully (see `/v1/demo` proof fields).
- Boundary: does not own deep CIJob (E-014) or full production metrics craft (E-008).
