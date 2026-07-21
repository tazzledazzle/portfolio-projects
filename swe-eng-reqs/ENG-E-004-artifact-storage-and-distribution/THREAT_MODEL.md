# Threat / failure notes — ENG-E-004

## Mitigations (Phase 3)

| ID | Threat | Mitigation |
|----|--------|------------|
| T-3-01 | Tampering via digest overwrite | `BlobStore.Put` / `putAt` reject PUT when digest exists and bytes differ (`ErrDigestConflict` → HTTP 409); identical re-upload is idempotent |
| T-3-03 | Path traversal in digest params | `ValidDigest` requires `sha256:` + 64 hex; rejects `..` |

## Other notes

- Authn/z: demo binds locally; production would require OIDC/RBAC.
- Multi-tenant isolation: enforce quotas before exposing shared endpoints.
- Supply chain: pin base images; sign artifacts in registry slices.
- Failure modes: dependency outage should degrade gracefully (see `/v1/demo` proof fields).
- Body size: `MaxBytesReader` capped at 16 MiB for PUT.
