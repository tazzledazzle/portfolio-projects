# ENG-H-004: Security and access control for developer systems

**Kind:** hidden | **Domain:** eng | **Stack:** go (stdlib only)

## Evidence from posting
Security and access control for developer systems

## What this proves
- **OIDC-inspired** claims (`iss` / `sub` / `aud` / `exp` / `roles`)
- **RBAC** allow/deny with default deny (`rbac_allow` / `rbac_deny`)
- Expired `exp` rejected
- Honesty: **`oidc_inspired`**, **`simulator`**, no external IdP (no Keycloak/Dex)

## Acceptance demo
```bash
make test
make demo-local
```

Port `:18704`. Proof: `oidc_inspired`, `rbac_allow`, `rbac_deny`, `simulator`.

## Endpoints
- `GET /healthz`, `GET /readyz`, `GET /v1/info`
- `POST /v1/auth/evaluate`
- `GET|POST /v1/demo`
- `GET /metrics`

## Does NOT own
Signing/SBOM (ENG-I-008) or an external OIDC IdP.
