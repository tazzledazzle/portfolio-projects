# Threat / failure notes — ENG-H-003

## Trust boundary
client → migration API (untrusted item payloads during v1→v2 dual-write)

## Mitigations (T-5-20 Tampering / dual-write split brain)
- Single mutex critical section writes both `v1` and `v2` maps when dual-write is enabled
- Compat tests (`TestMigrate_MidMigration_NoSplitBrain`) assert v2 never missing under `-race`
- Reject unsafe item IDs (path separators / traversal)
- Empty name rejected
- Body size limited via `http.MaxBytesReader` (1 MiB)

## Notes
- Authn/z: demo binds locally; production would gate migration controls
- Does not own OpenAPI/rate-limit craft (E-021) or IDP catalog (E-005)
- Field rename is intentional: `name` → `display_name` (Claude discretion)
