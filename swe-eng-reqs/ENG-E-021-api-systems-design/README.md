# ENG-E-021: API systems design

**Kind:** explicit | **Domain:** eng | **Stack:** go+compose

## Evidence from posting
and APIs

## Rationale
Public/internal API contracts for platforms.

## Ownership
Owns: OpenAPI contract craft, authz hooks, rate limits, compatibility tests.  
Does **not** own: IDP nouns (E-005), product SLAs (I-002), CLI goldens (I-007), dual-write migration (H-003).

## Honesty labels
- `openapi_inspired: true` — hand-authored OpenAPI 3.x YAML; **no** OAS codegen, gateway, or external IdP.
- Document served at `/v1/openapi.json` (YAML body) and present as `openapi.yaml`.

## Acceptance demo
Versioned OpenAPI service with authz, rate limits, and compatibility tests.

## Run

```bash
make test
make demo-local
make down
```

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-e-021:local`, apply with kubectl/Kind).

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info`
- `GET /v1/openapi.json`
- `GET /v1/resources` (requires `X-Scope: resources:read`)
- `GET /v2/resources` (requires `X-Scope: resources:read`)
- `GET|POST /v1/demo`
- `GET /metrics`

## Ports
`demo-local` listens on `:18521`.
