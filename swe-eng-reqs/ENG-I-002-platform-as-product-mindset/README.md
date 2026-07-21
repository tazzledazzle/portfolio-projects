# ENG-I-002: Platform-as-product mindset

**Kind:** implicit | **Domain:** eng | **Stack:** go+compose

## Evidence from posting
IDP + DevEx ownership + high-leverage

## Rationale
Success = developer adoption/experience, not feature count.

## Ownership
Owns: SLAs, adoption metrics, golden-path template.  
Does **not** own: IDP CRUD (E-005), ticket metrics (I-006), OpenAPI suite (E-021).

## Acceptance demo
Platform API with SLAs, adoption metrics endpoint, and golden-path template (`templates/golden-path.md`).

## Run

```bash
make test
make demo-local
make down
```

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-i-002:local`, apply with kubectl/Kind).

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info`
- `GET /v1/sla`
- `GET /v1/adoption`
- `GET /v1/golden-path`
- `GET|POST /v1/demo`
- `GET /metrics`

## Ports
`demo-local` listens on `:18602`.
