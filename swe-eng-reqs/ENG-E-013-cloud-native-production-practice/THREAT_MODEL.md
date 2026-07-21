# Threat / failure notes — ENG-E-013

| Threat | Disposition | Mitigation |
|--------|-------------|------------|
| T-5-07 Tampering of packaging flags | mitigate | Packaging facts derived from real Dockerfile/compose/deploy.yaml on disk — not client-supplied claims |
| Authn/z | accept (demo) | Demo binds locally; production would require OIDC/RBAC |
| Kind / cluster exposure | accept | Kind optional; demo-local is the phase gate |

## Failure modes
- Missing deploy.yaml → probes fall back to `/healthz`/`/readyz`; `hpa_ready` false.
- Client cannot assert `hpa_ready` via query params; only filesystem/manifest inspection counts.
