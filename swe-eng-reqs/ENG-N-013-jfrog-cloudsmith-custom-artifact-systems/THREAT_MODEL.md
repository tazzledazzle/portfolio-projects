# Threat / failure notes — ENG-N-013

## Trust boundary
client → custom-registry **simulator** API (untrusted tokens/scopes, artifact payloads, retention params)

## Threats (Phase 3)

| ID | Category | Mitigation |
|----|----------|------------|
| T-3-06 | Elevation of Privilege | Deny mutating routes without `artifacts:write` scope; return 401/403 |
| T-3-07 | Tampering (retention) | Retention deletes by keep-count only; never rewrite digest bytes on remaining artifacts |
| T-3-08 | Spoofing (fake scanner / vendor trust) | Scan returns **fixture** findings only; `simulator: true` + `vendor_model: custom-registry` in `/v1/info` and demo; README SIMULATOR block |
| T-3-SC | Tampering (supply chain) | Stdlib only — no JFrog/Cloudsmith SDKs in go.mod |

## Notes
- This is a **simulator**, not a live Artifactory/Cloudsmith integration.
- Production would require real OIDC/RBAC and a real scanner product — out of scope for this portfolio MVP.
