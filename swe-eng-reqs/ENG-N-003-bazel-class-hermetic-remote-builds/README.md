# ENG-N-003: Bazel-class hermetic/remote builds

**Kind:** nice | **Domain:** eng | **Stack:** go+compose

## Evidence from posting
Bazel

## Rationale
Hermetic/remote build systems for large codebases.

## Acceptance demo
Content-addressed remote cache with hermetic input digests and hit metrics.

## Run

```bash
make test
make demo
make down
```

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-n-003:local`, apply with kubectl/Kind).

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info`
- `GET|POST /v1/demo`
- `GET /metrics`
