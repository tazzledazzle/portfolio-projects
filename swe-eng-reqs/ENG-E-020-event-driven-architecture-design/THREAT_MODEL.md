# Threat / failure notes — ENG-E-020

## Trust boundary
Untrusted clients → `/v1/publish`, `/v1/consume`, `/v1/replay` (envelopes + offsets).

## STRIDE mitigations (plan T-5-09, T-5-SC)

| Threat | Mitigation |
|--------|------------|
| T-5-09 Tampering / replay double-apply | Replay uses explicit `from_offset` cursor; consumers Ack/Fail separately; demo does not auto-reapply side effects |
| T-5-SC Supply chain / fake NATS | Stdlib only; no `nats-io/nats.go`; honesty labels `nats_inspired` + `simulator` |
| DoS via oversized body | `http.MaxBytesReader` 1 MiB on write endpoints |
| Authn/z | Demo binds locally; production would require OIDC/RBAC |

## Failure modes
- Handler `Fail` appends to DLQ with reason; message remains in append-only log for replay.
- Replay is offline from in-memory log — never claims JetStream or live NATS connectivity.
