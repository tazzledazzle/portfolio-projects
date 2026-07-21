# Threat / failure notes — ENG-H-001

## Trust boundary
client → chaos / blast-radius API (untrusted domain/scenario inputs)

## Mitigations (T-5-13 Tampering)
- Blast radius affected/unaffected sets are **server-computed** from scenario state — clients cannot supply containment claims
- Reject unsafe domain/tenant IDs (path separators / traversal)
- Chaos requires a registered domain; unknown domains denied
- Empty scenario rejected

## Notes
- Authn/z: demo binds locally; production would gate chaos injection
- Simulator / in-process only — not live cluster chaos tooling
- Body size limited via `http.MaxBytesReader` (1 MiB)
- Does not own E-010 fan-out or I-004 quota enforcement
