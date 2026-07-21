# ENG-E-025: Ephemeral development environments

**Kind:** explicit | **Domain:** eng | **Stack:** go+k8s

## Evidence from posting
ephemeral development environments

## Rationale
Explicit platform surface for short-lived envs.

## Acceptance demo
DevEnv CRD with TTL reclaim and Ready/Expired conditions.

## Run

```bash
make test
make demo-local # in-memory TTL reconcile proof on :18325 (phase gate)
make demo
make down
```

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-e-025:local`, apply with kubectl/Kind).
The DevEnv CRD can optionally be installed with
`kubectl apply -f k8s/crd.yaml`; Kind and kubeconfig are not required by
`make test` or `make demo-local`.

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info`
- `POST /v1/devenvs` (`{"id":"preview","ttl_seconds":60}`)
- `GET /v1/devenvs/{id}`
- `POST /v1/devenvs/{id}/tick` (logical-clock demo helper)
- `POST /v1/reconcile`
- `GET|POST /v1/demo`
- `GET /metrics`

Creation sets `Ready=True` and `Expired=False`. Once the TTL elapses, reconcile
sets `Ready=False`, `Expired=True`, and `reclaimed=true`.
