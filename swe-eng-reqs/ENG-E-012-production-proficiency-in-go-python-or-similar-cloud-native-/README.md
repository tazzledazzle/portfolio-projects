# ENG-E-012: Production proficiency in Go, Python, or similar cloud-native languages

**Kind:** explicit | **Domain:** eng | **Stack:** go+compose+k8s

## Evidence from posting
Proficiency in Go, Python, or similar languages with production experience shipping services in cloud-native environments

## Rationale
**OR semantics (CLAUDE.md / D-08):** the requirement is Go **or** Python **or** similar —
not both. This folder proves proficiency with a **single Go** production-style
`net/http` service. No Python package is required or included.

## Acceptance demo
Single Go control-plane sample with compose and K8s deploy. Live `/v1/demo` on
port **18512** returns `language=go`, `or_semantics=true`, `production_sample=true`.

## Run

```bash
make test
make demo-local
make demo
make down
```

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-e-012:local`, apply with kubectl/Kind).

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info` — includes `language`, `or_semantics`
- `GET|POST /v1/demo`
- `GET /metrics`
