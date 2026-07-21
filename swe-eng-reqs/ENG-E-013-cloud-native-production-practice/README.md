# ENG-E-013: Cloud-native production practice

**Kind:** explicit | **Domain:** eng | **Stack:** go+compose+k8s

## Evidence from posting
production experience shipping services in cloud-native environments

## Rationale
Owns **Dockerfile + probes + HPA-ready + compose parity**. Does not own IDP catalog, queue/workflow, or operator reconcile. Kind is optional; `make demo-local` is the gate.

## Acceptance demo
Live `/v1/demo` on port **18513** proves `probes`, `hpa_ready`, `compose_parity`, and `dockerfile` from real files/manifests (not client claims).

## Run

```bash
make test
make demo-local
make demo   # optional Docker Compose
make down
```

Kubernetes: `k8s/deploy.yaml` includes liveness/readiness probes and a `HorizontalPodAutoscaler` (Kind optional).

## Endpoints
- `GET /healthz` — liveness
- `GET /readyz` — readiness
- `GET /v1/info`
- `GET /v1/packaging` — packaging facts from Dockerfile/compose/deploy
- `GET|POST /v1/demo` — live packaging proof
- `GET /metrics`
