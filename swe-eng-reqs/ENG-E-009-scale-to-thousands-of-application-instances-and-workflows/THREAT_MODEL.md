# Threat / failure notes — ENG-E-009

## Trust boundary
Untrusted clients → `/v1/simulate` (count) and `/v1/demo`.

## STRIDE mitigations (plan T-5-10)

| Threat | Mitigation |
|--------|------------|
| T-5-10 Denial of Service via huge simulate | Cap `maxSimulate` (50k); reject oversized counts; backpressure when queue capacity saturated |
| Authn/z | Demo binds locally; production would require OIDC/RBAC |
| Confusion with durable workflows | No signal/durable/replay APIs; `durable_workflow: false` in Info |

## Failure modes
- When queue capacity is hit, admissions are rejected or delayed (`backpressure: true`).
- Latency samples drive `p99_ms` for SLO-style reporting — not a production HPA controller.
