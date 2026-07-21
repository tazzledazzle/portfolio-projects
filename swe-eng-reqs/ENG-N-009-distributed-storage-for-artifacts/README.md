# ENG-N-009: Distributed storage for artifacts

**Kind:** nice | **Domain:** eng | **Stack:** go+compose

## Evidence from posting
distributed storage, caching, versioning, or multi-region artifact distribution

## Rationale
Storage-system depth behind registries.

## Acceptance demo
Multi-node object storage facade with durability metadata.

## Run

```bash
make test
make demo-local # in-process 3-node proof on :18309
make demo
make down
```

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-n-009:local`, apply with kubectl/Kind).

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info`
- `PUT /v1/objects` (raw object bytes; returns `sha256:` digest and durability)
- `GET /v1/objects/{digest}` (bytes plus durability headers)
- `GET /v1/durability/{digest}` (`replicas`, `checksum`, `healthy_nodes`, `quorum`)
- `GET|POST /v1/demo`
- `GET /metrics`

The local facade writes each object to all healthy in-process nodes and only
acknowledges the write when the configured quorum is available. It reports
replica and healthy-node counts separately so the demo never overstates
durability.
