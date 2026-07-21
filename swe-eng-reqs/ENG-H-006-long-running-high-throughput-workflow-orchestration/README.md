# ENG-H-006: Long-running / high-throughput workflow orchestration

**Kind:** hidden | **Domain:** eng | **Stack:** go+compose (stdlib only)

## Evidence from posting
thousands of workflows + delivery/CI ownership

## Rationale
Workflow scale implies queues, retries, durable state — proven here as a durable multi-step engine MVP with measured throughput (not a load-only sim).

## Boundary (D-03)
| Owns | Does NOT own |
|------|--------------|
| Durable workflow MVP + throughput numbers | E-009 load-only Simulate; E-019 queue partitions; I-009 GPU chunk upload |

## Acceptance demo
Durable workflow engine MVP with `throughput_per_s`, `steps_completed`, `replay_safe` on `:18706`.

## Run

```bash
make test
make demo-local   # port 18706
make demo         # optional Docker
make down
```

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-h-006:local`, apply with kubectl/Kind).

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info`
- `POST /v1/workflows` — start durable multi-step workflow
- `POST /v1/workflows/{id}/signal` — advance step (idempotent by `event_id`)
- `GET|POST /v1/demo` — live proof: durable, throughput_per_s, steps_completed, replay_safe
- `GET /metrics`
