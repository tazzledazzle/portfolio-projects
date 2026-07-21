# Threat / failure notes — ENG-E-003

## T-4-01 Unauthorized promote / rollback (Elevation of Privilege)

- **Mitigation (MVP):** Validate release IDs; append an audit entry on every mutate (`promote`, `rollback`). Demo is local-only (`demo-local` on :18403).
- **Deferred:** Production OIDC/RBAC for delivery APIs → Phase 5 (ENG-H-004).

## Other notes

- Authn/z: demo binds locally; production would require OIDC/RBAC.
- Multi-tenant isolation: enforce quotas before exposing shared endpoints.
- Supply chain: pin base images; sign artifacts in registry slices.
- Failure modes: unknown release IDs return errors; concurrent promote is mutex-safe.
