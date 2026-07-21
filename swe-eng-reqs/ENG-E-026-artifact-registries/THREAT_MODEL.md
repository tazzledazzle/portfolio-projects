# Threat / failure notes — ENG-E-026

## Mitigations (Phase 3)

| ID | Threat | Mitigation |
|----|--------|------------|
| T-3-02 | Tag hijack / digest overwrite confusion | Separate `tags` map from `manifests`; retarget mutates pointer only; validate digest grammar before `PutTag` |
| T-3-03 | Path traversal in name/tag | Reject `".."` in repository and tag names (`ErrUnsafeName`) |
| T-3-04 | Spoofing as full OCI Distribution | README + `/v1/info` label **OCI-inspired MVP (not conformance-tested)**; `conformance: false` |

## Other notes

- Authn/z: demo binds locally; production would require OIDC/RBAC.
- Multi-tenant isolation: enforce quotas before exposing shared endpoints.
- Supply chain: pin base images; sign artifacts in registry slices.
- Failure modes: dependency outage should degrade gracefully (see `/v1/demo` proof fields).
- Resumable blob upload sessions intentionally out of scope (monolithic PUT only).
