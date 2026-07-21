# Threat / failure notes — ENG-E-014

| Threat | Disposition | Mitigation |
|--------|-------------|------------|
| T-5-16 Condition spoof | mitigate | Complete/Failed computed server-side from Job status in `Reconcile`; clients cannot set conditions via API |
| Authn/z | accept (demo) | Local bind only; production would require cluster RBAC |
| Kind/apiserver dependency | mitigate | Kind optional; `demo-local` in-memory is the gate |

- Multi-tenant isolation: enforce quotas before exposing shared endpoints.
- Failure modes: pending Jobs stay Ready; terminal paths clear Ready.
