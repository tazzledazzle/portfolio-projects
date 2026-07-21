# Threat / failure notes — ENG-E-021

## Trust boundary
client → versioned API (untrusted scopes, high-volume requests)

## Mitigations (T-5-03)
- **Denial of Service / rate-limit bypass:** per-subject counters enforced server-side in `Allow`; exceed → `429 rate_limited`.
- **Request size:** handlers wrap body with `http.MaxBytesReader` (1 MiB).
- **Authz default-deny:** missing/wrong scopes → 403; required scope `resources:read`.
- Demo binds locally (`demo-local` on `:18521`).

## Out of scope
- IDP catalog CRUD (ENG-E-005)
- Dual-write migration craft (ENG-H-003)
- External OIDC IdP
