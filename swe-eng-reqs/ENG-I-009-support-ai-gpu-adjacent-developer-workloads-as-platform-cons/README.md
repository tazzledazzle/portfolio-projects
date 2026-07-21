# ENG-I-009: Support AI/GPU-adjacent developer workloads as platform consumers

**Kind:** implicit | **Domain:** eng | **Stack:** go+compose (stdlib only)

## Evidence from posting
CoreWeave AI cloud + DevEx serving that org

## Rationale
Platforms serve AI engineers: large artifacts, long jobs — not GPU kernel expertise.

## Boundary (D-03)
| Owns | Does NOT own |
|------|--------------|
| Long-running job + chunked large-artifact upload with timeout | H-006 durable workflow engine; E-013 HPA packaging; GPU kernel scheduling |

## Chunk policy (Claude discretion)
- Default max chunk: **64KiB**
- Default total cap: **8MiB** (above typical `MaxBytesReader` 1<<20 for multi-chunk demos)

## Acceptance demo
Long-running job + chunked upload with timeout on `:18609` — proof fields `long_running`, `timeout`, `chunked_upload`, `bytes_received`.

## Run

```bash
make test
make demo-local   # port 18609
make demo         # optional Docker
make down
```

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-i-009:local`, apply with kubectl/Kind).

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info`
- `POST /v1/jobs` — start long-running job (`timeout_ms` optional)
- `POST /v1/jobs/{id}/chunks` — upload artifact chunk (body = bytes)
- `GET|POST /v1/demo` — live proof: long_running, timeout, chunked_upload, bytes_received
- `GET /metrics`
