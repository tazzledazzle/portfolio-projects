# ENG-E-005: Internal developer platforms

**Kind:** explicit | **Domain:** eng | **Stack:** go+compose

## Evidence from posting
internal developer platforms

## Rationale
Platform-as-product for internal engineers — catalog nouns only (Boundary Matrix).

## Ownership
Owns: projects / pipelines / environments self-service catalog.  
Does **not** own: OpenAPI craft (E-021), product SLAs (I-002), ticket metrics (I-006), CLI (I-007).

## Acceptance demo
Self-service platform API for projects, pipelines, and environments with live `/v1/demo` proof.

## Run

```bash
make test
make demo-local
make down
```

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-e-005:local`, apply with kubectl/Kind).

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info`
- `POST|GET /v1/projects`
- `POST|GET /v1/pipelines`
- `POST|GET /v1/environments`
- `GET|POST /v1/demo`
- `GET /metrics`

## Ports
`demo-local` listens on `:18505`.
