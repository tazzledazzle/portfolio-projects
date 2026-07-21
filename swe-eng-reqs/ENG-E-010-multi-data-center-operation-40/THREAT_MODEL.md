# Threat / failure notes — ENG-E-010

## Trust boundary
client → multi-DC simulator API (untrusted DC IDs / fan-out payloads)

## Mitigations (T-5-12)
- Cap registered DCs (`maxDataCenters`) to bound DoS via registration flood
- Reject unsafe IDs (path separators / traversal)
- Fan-out skips unhealthy DCs and reports partial success (`fanout_ok` + pushed/failed counts)
- Domain health checked before config push

## Notes
- Authn/z: demo binds locally; production would require OIDC/RBAC
- Simulator only — not physical multi-DC blast-radius engineering (see ENG-H-001)
- Body size limited via `http.MaxBytesReader` (1 MiB)
