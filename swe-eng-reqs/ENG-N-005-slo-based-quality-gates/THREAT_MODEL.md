# Threat / failure notes — ENG-N-005

## T-4-03 Evidence forgery (Spoofing)

- **Mitigation:** Evidence (`burn_short`, `burn_long`) is computed server-side from ingested series. Client-supplied `burn_rate` on `/v1/gates/evaluate` is ignored.
- Honesty labels: `promql_inspired` + `simulator` on `/v1/info`, README, and demo proof (does NOT connect to Prometheus).

## Other notes

- Authn/z: demo binds locally; production would require OIDC/RBAC.
- Multi-tenant isolation: enforce quotas before exposing shared endpoints.
- Supply chain: pin base images; sign artifacts in registry slices.
- Failure modes: unknown SLO IDs return errors; concurrent ingest/evaluate is mutex-safe.
