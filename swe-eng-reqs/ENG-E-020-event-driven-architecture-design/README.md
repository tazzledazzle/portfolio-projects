# ENG-E-020: Event-driven architecture design

**Kind:** explicit | **Domain:** eng | **Stack:** go+compose (stdlib only)

## Evidence from posting
event-driven architectures

## Rationale
Distinct from sync request/response APIs.

## Honesty labels (D-06, D-10, D-11)
This slice is a **NATS-inspired** in-memory **simulator**. It does **not** connect to a NATS server or JetStream.
- `nats_inspired: true`
- `simulator: true`
- `nats_connected: false`
- No `nats-io/nats.go` dependency

## Acceptance demo
In-process event bus with schema envelopes, consumers, DLQ, and replay from log offset.

## Boundary (D-03)
- **Owns:** envelopes, consumers, DLQ, replay
- **Does not own:** queue idempotency (ENG-E-019), durable workflow engine (ENG-H-006), REST OpenAPI craft (ENG-E-021)

## Run

```bash
make test
make demo-local   # :18520
make demo
make down
```

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-e-020:local`, apply with kubectl/Kind).

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info` — honesty: `nats_inspired`, `simulator`
- `POST /v1/publish` — `{subject, schema, payload}`
- `POST /v1/consume` — optional `{consumer, fail}` (fail → DLQ)
- `POST /v1/replay` — `{from_offset}`
- `GET|POST /v1/demo` — live proof: nats_inspired, simulator, dlq, replay, schema_envelope
- `GET /metrics`
