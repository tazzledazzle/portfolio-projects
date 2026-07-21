# ENG-I-006: Automation-first / self-service

**Kind:** implicit | **Domain:** eng | **Stack:** go+compose

## Evidence from posting
automation-first / self-service

## Rationale
Remove ticket bottlenecks with measurable before/after automation metrics.

## Ownership
Owns: ticket→self-service conversion metrics (`ticket_removed`, before/after counts).  
Does **not** own: full IDP catalog (E-005), product SLAs (I-002), CLI UX (I-007).

## Acceptance demo
Self-service request path proves `ticket_removed` with server-computed automation metrics.

## Run

```bash
make test
make demo-local
make down
```

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-i-006:local`, apply with kubectl/Kind).

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info`
- `POST /v1/requests`
- `GET /v1/metrics`
- `GET|POST /v1/demo`
- `GET /metrics`

## Ports
`demo-local` listens on `:18606`.
