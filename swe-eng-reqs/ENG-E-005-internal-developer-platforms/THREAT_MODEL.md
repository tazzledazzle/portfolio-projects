# Threat / failure notes — ENG-E-005

## Trust boundary
client → IDP catalog API (untrusted project/pipeline/environment names)

## Mitigations (T-5-01)
- **Elevation of Privilege / path injection:** `Create*` rejects names containing `..`, `/`, or `\`; empty names rejected.
- **Unauthorized catalog mutate:** demo binds locally (`demo-local` on `:18505`); production would require OIDC/RBAC before exposing mutate routes.
- **Input size:** handlers wrap body with `http.MaxBytesReader` (1 MiB).

## Out of scope
- OpenAPI rate limits (ENG-E-021)
- Product SLAs / adoption (ENG-I-002)
- Ticket-removal metrics (ENG-I-006)
