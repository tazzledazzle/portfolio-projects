# ENG-E-009: Scale to thousands of application instances and workflows

**Kind:** explicit | **Domain:** eng | **Stack:** go+compose

## Evidence from posting
reliably serve thousands of application instances and engineering workflows

## Rationale
Throughput/concurrency at workflow/instance scale is explicit.

## Acceptance demo
Load simulation of **≥1000** workflows with queueing, **backpressure**, and SLO latency (`p99_ms`, `queue_depth`).

## Boundary (D-03)
- **Owns:** ≥1000 workflow load, backpressure, SLO latency metrics
- **Does not own:** durable workflow engine (ENG-H-006), queue idempotency (ENG-E-019), multi-DC topology (ENG-E-010)

## Run

```bash
make test
make demo-local   # :18509
make demo
make down
```

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-e-009:local`, apply with kubectl/Kind).

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info`
- `POST /v1/simulate` — `{count}` (capped; default 1000)
- `GET /v1/metrics` — JSON `p99_ms`, `queue_depth`
- `GET|POST /v1/demo` — live proof: workflows_simulated≥1000, backpressure, p99_ms, queue_depth
- `GET /metrics` — Prometheus text
