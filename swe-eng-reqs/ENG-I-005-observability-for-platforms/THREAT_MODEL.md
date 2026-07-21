# Threat / failure notes — ENG-I-005

- Authn/z: demo binds locally; production would require OIDC/RBAC.
- Span input: names are bounded and validated before entering the in-memory store.
- T-4-09 (alert rule injection / tampering): clients select only compiled-in
  fixture rules and provide numeric samples; no expressions or code are
  evaluated.
- T-4-10 (false vendor claim / spoofing): README, `/v1/info`, trace exports,
  and demo proofs label the service `otel_inspired`, `simulator`, and
  `collector: none`.
- Scope: stdlib-only in-process evidence; no OTel SDK, collector, Tempo, or
  Grafana connection and no runbook library or release-gate decisions.
