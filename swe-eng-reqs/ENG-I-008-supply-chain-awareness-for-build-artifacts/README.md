# ENG-I-008: Supply-chain awareness for build/artifacts

**Kind:** implicit | **Domain:** eng | **Stack:** go (stdlib crypto only)

## Evidence from posting
Supply-chain awareness for build artifacts

## What this proves
- **ed25519** sign/verify of content digests (`crypto/ed25519`)
- **SPDX-inspired** SBOM document fields (`spdx_inspired: true`)
- Registry **push scopes** with default deny (`artifacts:push`)
- Honesty: **`sigstore: false`** — no cosign/Sigstore

## Fixture keys
Demo keys live under `testdata/keys/` only. Private keys are never logged.

## Acceptance demo
```bash
make test
make demo-local
```

Port `:18608`. Proof: `signed`, `sbom_spdx_inspired`, `scope_enforced`, `sigstore=false`.

## Endpoints
- `GET /healthz`, `GET /readyz`, `GET /v1/info`
- `POST /v1/sign`, `POST /v1/sbom`, `POST /v1/push`
- `GET|POST /v1/demo`
- `GET /metrics`

## Does NOT own
Full OCI registry (Phase 3), OIDC/RBAC engine (ENG-H-004), or Sigstore.
