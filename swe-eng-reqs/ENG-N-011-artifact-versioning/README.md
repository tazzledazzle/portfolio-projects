# ENG-N-011: Artifact versioning

**Kind:** nice | **Domain:** eng | **Stack:** go+compose

## Evidence from posting
versioning

## Rationale
Immutability, tags, promotion of versions.

## Acceptance demo
Version/promotion API with immutable digests and mutable tags.
Promote advances `dev → staging → prod` without changing the version digest; tags may retarget independently.

## Run

```bash
make test
make demo-local
make demo   # requires Docker
make down
```

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-n-011:local`, apply with kubectl/Kind).

`demo-local` listens on **:18311**.

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info`
- `GET|POST /v1/demo` — proof: `promoted`, `digest_unchanged`, `tag_mutable`
- `POST /v1/versions` — create `{name, digest, stage}`
- `POST /v1/versions/{id}/promote` — advance one stage; digest unchanged
- `PUT /v1/tags/{tag}` — mutable tag → digest
- `GET /v1/tags/{tag}`
- `GET /metrics`
