# Threat / failure notes — ENG-I-009

## Trust boundary
client → I-009 API (untrusted chunk bodies)

## STRIDE mitigations (plan T-5-15)

| Threat | Mitigation |
|--------|------------|
| T-5-15 DoS — oversized upload | Per-chunk cap (`DefaultMaxChunkBytes` 64KiB) + total artifact cap (8MiB); `http.MaxBytesReader` on chunk route; job timeout rejects late chunks |
| Authn/z | Demo binds locally; production would require OIDC/RBAC |
| Multi-tenant isolation | Enforce quotas before exposing shared endpoints |
| Supply chain | Stdlib only (D-02); pin base images in compose/k8s |
| Failure modes | Unknown job → 404; timed-out job → 410 Gone; oversize → 413 |

## Out of scope
Durable workflow Signal API (H-006), HPA packaging (E-013), GPU kernel scheduling.
