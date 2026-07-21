# ENG-E-004: Artifact storage and distribution

**Kind:** explicit | **Domain:** eng | **Stack:** go+compose

## Evidence from posting
artifact storage and distribution

## Rationale
Binary/package/image lifecycle is core DevEx.

## Acceptance demo
Push/pull content-addressed blobs with digest immutability and metadata.

## Digests
All digests use OCI form: `sha256:` + 64 lowercase hex characters.

## Run

```bash
make test
make demo-local   # port 18304, no Docker
make demo         # compose
make down
```

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-e-004:local`, apply with kubectl/Kind).

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info`
- `GET|POST /v1/demo` — live proof: `digest_immutable`, `blob_count`, `metadata_keys`
- `PUT /v1/blobs` — raw body; optional metadata via `X-Meta-*` headers (e.g. `X-Meta-Content-Type`); returns `{digest, size, metadata}` (201)
- `PUT /v1/blobs/{digest}` — client-addressed put; **409** if digest exists with different bytes
- `GET /v1/blobs/{digest}` — blob bytes
- `HEAD /v1/blobs/{digest}` — metadata / content-length
- `GET /metrics`

## Boundary
Owns immutable blob PUT/GET by digest + metadata. Does **not** own tag APIs, multi-region, or retention.
