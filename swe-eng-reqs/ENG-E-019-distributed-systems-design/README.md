# ENG-E-019: Distributed systems design

**Kind:** explicit | **Domain:** eng | **Stack:** go+compose

## Evidence from posting
You think in distributed services

## Rationale
Core design competency for CoreWeave-scale platforms.

## Acceptance demo
Distributed task queue with retries, idempotency keys, and partition handling.

## Boundary (D-03)
- **Owns:** task queue retries, idempotency keys, partitions, `duplicate_suppressed`
- **Does not own:** event bus / DLQ / replay (ENG-E-020), durable workflows (ENG-H-006), load/backpressure SLO sim (ENG-E-009)

## Run

```bash
make test
make demo-local   # :18519
make demo         # Docker Compose
make down
```

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-e-019:local`, apply with kubectl/Kind).

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info`
- `POST /v1/tasks` — enqueue `{payload, idempotency_key, partition}`
- `POST /v1/ack` — `{id}`
- `POST /v1/nack` — `{id}` (increments attempts, requeues)
- `GET|POST /v1/demo` — live proof: idempotent, duplicate_suppressed, partitions, retries
- `GET /metrics`
