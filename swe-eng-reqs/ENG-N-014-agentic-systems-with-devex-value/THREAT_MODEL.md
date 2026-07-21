# Threat / failure notes — ENG-N-014

- Authn/z: demo binds locally; production would require OIDC/RBAC.
- Multi-tenant isolation: enforce quotas before exposing shared endpoints.
- Supply chain: pin base images; sign artifacts in registry slices.
- Failure modes: dependency outage should degrade gracefully (see `/v1/demo` proof fields).

| ID | Threat | Mitigation |
|---|---|---|
| T-6-10 | Fixture results are misrepresented as production telemetry | Server owns provenance fields and always emits `baseline_source=fixture` and `fabricated_prod=false`. |
| T-6-10A | A client attempts to override honesty fields | POST bodies are not used as metric or provenance inputs; the server reloads its on-disk fixture. |
| T-6-10B | Invalid durations create misleading percentages | ROI calculation rejects non-positive fixture durations. |

Production ROI would require instrumented adoption, quality, and incident
measurements with documented sampling and confidence intervals.
