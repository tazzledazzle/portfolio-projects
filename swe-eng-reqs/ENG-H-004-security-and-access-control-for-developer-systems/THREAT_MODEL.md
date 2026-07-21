# Threat / failure notes — ENG-H-004

| Threat | Disposition | Mitigation |
|--------|-------------|------------|
| T-5-19 Authz bypass | mitigate | Default deny; validate `exp`; policy tests cover allow and deny |
| External IdP trust | mitigate | Simulator only — no Keycloak/Dex/OIDC client; `external_idp=false` |
| T-5-SC IdP deps | mitigate | Stdlib only |

- Claims are accepted as JSON input for the demo evaluator (not verified JWTs).
- Production would verify signatures against a real IdP JWKS (out of scope).
