# ENG-E-026: Artifact registries

**Kind:** explicit | **Domain:** eng | **Stack:** go+compose

## Important: OCI-inspired MVP (not conformance-tested)

This service is an **OCI-inspired MVP / simulator**. It demonstrates tag→digest resolve and
retarget semantics for portfolio evidence. It is **not** OCI Distribution Spec
conformance-tested and does **not** connect to a production registry vendor.

## Evidence from posting
artifact registries

## Rationale
Registry as a concrete system, not only storage.

## Acceptance demo
Tag resolve returns digest; retargeting a tag does not change stored manifest bytes.

## Digests
All digests use OCI form: `sha256:` + 64 lowercase hex characters.

## Run

```bash
make test
make demo-local   # port 18326, no Docker
make demo
make down
```

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-e-026:local`, apply with kubectl/Kind).

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info` — includes `oci_inspired: true`, `conformance: false`
- `GET|POST /v1/demo` — proof: `tag_to_digest`, `tag_mutable`, `digest_immutable`
- `PUT /v1/registry/{name}/manifests` — store manifest bytes (monolithic PUT; no resumable upload sessions)
- `GET /v1/registry/{name}/manifests/{digest}`
- `PUT /v1/registry/{name}/tags/{tag}` — JSON `{"digest":"sha256:..."}`
- `GET /v1/registry/{name}/tags/{tag}` — resolve tag→digest
- `GET /metrics`

## Boundary
Owns tag→digest registry resolve/retarget. Does **not** own multi-node durability, promotion, or scan.
