# Threat / failure notes — ENG-E-024

## T-4-02 Stuck canary / DoS via Step (Tampering)

- **Mitigation:** `Abort` forces weight 0 and status `aborted`; terminal `aborted`/`promoted` reject further `Step`.
- Local-only demo on :18424.

## Other notes

- Authn/z: demo binds locally; production would require OIDC/RBAC.
- Multi-tenant isolation: enforce quotas before exposing shared endpoints.
- Supply chain: pin base images; sign artifacts in registry slices.
- Failure modes: unknown canary IDs return errors; concurrent Step is mutex-safe.
