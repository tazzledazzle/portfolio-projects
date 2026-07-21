# Threat / failure notes — ENG-E-019

## Trust boundary
Untrusted clients → `/v1/tasks`, `/v1/ack`, `/v1/nack` (payload + idempotency keys).

## STRIDE mitigations (plan T-5-08)

| Threat | Mitigation |
|--------|------------|
| T-5-08 Tampering / duplicate apply | Idempotency keys suppress duplicate side effects (`duplicate_suppressed`); same key returns original task id |
| DoS via oversized body | `http.MaxBytesReader` 1 MiB on write endpoints |
| Authn/z | Demo binds locally; production would require OIDC/RBAC |
| Multi-tenant isolation | Enforce quotas before exposing shared queue endpoints |

## Failure modes
- Nack requeues with incremented `attempts`; callers must Ack to complete.
- Partition hint is modulo partition count — invalid negative hints clamp to 0.
