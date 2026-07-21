# Threat / failure notes — ENG-H-006

## Trust boundary
client → H-006 API (untrusted workflow signals)

## STRIDE mitigations (plan T-5-14)

| Threat | Mitigation |
|--------|------------|
| T-5-14 Tampering — replay double-apply | Signal is idempotent by `(workflow_id, event_id)`; duplicate events do not advance `steps_completed`; demo proves `replay_safe` |
| Authn/z | Demo binds locally; production would require OIDC/RBAC |
| Multi-tenant isolation | Enforce quotas before exposing shared endpoints |
| Supply chain | Stdlib only (D-02); pin base images in compose/k8s |
| Failure modes | Unknown workflow → 404; complete workflow rejects further advances |

## Out of scope
Load-only Simulate (E-009), queue partitions (E-019), GPU chunked upload (I-009).
