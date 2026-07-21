# ENG-N-013: JFrog / Cloudsmith / custom artifact systems

**Kind:** nice | **Domain:** eng | **Stack:** go+compose

## Important: This is a SIMULATOR

This service simulates a **custom artifact registry** (scopes, retention, scan-hook stub) for portfolio demonstration.

It does **NOT** connect to JFrog Artifactory, Cloudsmith, or any vendor API.
It does **NOT** use vendor SDKs or credentials.
It is **NOT** production-ready.

Purpose: Demonstrate understanding of commercial/custom registry policy surfaces (auth scopes, retention, vulnerability-scan hooks) without claiming live vendor integration.

## Evidence from posting
e.g. JFrog Artifactory, Cloudsmith, or custom-built systems

## Rationale
Commercial or custom registry experience.

## Acceptance demo
Custom registry **simulator** with retention, auth scopes, and vulnerability-scan hook stub (fixture findings only).

## Run

```bash
make test
make demo-local
make demo   # requires Docker
make down
```

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-n-013:local`, apply with kubectl/Kind).

`demo-local` listens on **:18313**.

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info` — includes `"simulator": true`, `"vendor_model": "custom-registry"`
- `GET|POST /v1/demo` — proof: `retention_deleted`, `auth_scope_enforced`, `scan_hook`, `simulator`
- `PUT /v1/artifacts` — requires `Authorization: Bearer demo` and/or `X-Scope: artifacts:write`
- `POST /v1/retention/run` — body `{ "keep": N }`; deletes older artifacts by count
- `POST /v1/scan` — scan-hook **stub**; returns fixture findings (no network)
- `GET /metrics`
