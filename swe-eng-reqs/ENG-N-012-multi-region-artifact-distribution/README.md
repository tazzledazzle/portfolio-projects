# ENG-N-012: Multi-region artifact distribution

**Kind:** nice | **Domain:** eng | **Stack:** go+compose

## Evidence from posting
multi-region artifact distribution

## Rationale
Latency/availability for global consumers.

## Acceptance demo
Async replication across region buckets with lag metrics.

## Run

```bash
make test
make demo-local # async in-memory proof on :18312
make demo
make down
```

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-n-012:local`, apply with kubectl/Kind).

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info`
- `PUT /v1/regions/{region}/blobs` (raw bytes; returns `sha256:` digest)
- `GET /v1/regions/{region}/blobs/{digest}` (404 while replication is pending)
- `GET /v1/replication/status` (`regions`, `pending`, `lag_ms`)
- `GET|POST /v1/demo`
- `GET /metrics`

The primary write returns before background fan-out. `lag_ms` remains positive
while a secondary read is unavailable and reaches zero only after all regional
copies complete. The live demo reports both the pre-sync lag and successful
secondary read.
