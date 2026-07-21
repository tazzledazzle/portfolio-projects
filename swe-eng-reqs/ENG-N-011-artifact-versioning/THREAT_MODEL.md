# Threat / failure notes — ENG-N-011

## Trust boundary
client → version/promote/tag API (untrusted JSON, tag names, digests)

## Threats (Phase 3)

| ID | Category | Mitigation |
|----|----------|------------|
| T-3-05 | Tampering (tag poisoning vs digest) | Promote mutates **stage only**; `Digest` field is immutable; tags live in a separate map from version digests |
| T-3-09 | Tampering (path/digest) | Reject `..` in names; normalize digests to `sha256:<64hex>`; reject unnormalized garbage |

## Notes
- Authn/z: demo binds locally; production would require OIDC/RBAC.
- This slice does **not** own object replica topology or regional lag (see ENG-N-009 / ENG-N-012).
