# Threat / failure notes — ENG-I-004

## Trust boundary
client → tenant schedule/quota API (untrusted tenant IDs and unit counts)

## Mitigations (T-5-11 Elevation of Privilege / cross-tenant steal)
- Require non-empty tenant ID on SetQuota and Schedule (default deny without identity)
- Reject unsafe tenant IDs (path separators / traversal)
- Per-tenant quota isolation — schedule beyond quota returns `ErrQuotaExceeded`
- Noisy-neighbor rate limits isolate high-rate tenants without blocking others
- Unknown tenant without configured quota is denied (default deny)

## Notes
- Authn/z: demo binds locally; production would map identity → tenant_id via OIDC/RBAC
- Body size limited via `http.MaxBytesReader` (1 MiB)
- Does not own multi-DC topology or chaos blast-radius proofs
