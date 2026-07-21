# ENG-N-008: Heterogeneous deployment environments

**Kind:** nice | **Domain:** eng | **Stack:** go+compose

## Evidence from posting
heterogeneous deployment environments

## Rationale
One workload, many runtimes: schedule the same workload across a mix of
Kubernetes and VM deployment profiles.

## Scope (Boundary Matrix D-03 / D-11)

- **Owns:** the profile abstraction and scheduling the **same workload** across
  **≥3 heterogeneous profiles**.
- **Profiles (D-11):**
  - `k8s-standard` — Kubernetes, no GPU
  - `k8s-gpu` — Kubernetes with GPU constraint
  - `vm-bake` — VM image bake runtime
- **Does NOT own:** canary weights, burn-rate gates, or reconcile finalizers.

## Acceptance demo
Register the three profiles, then schedule one workload across all of them.
`/v1/demo` proves `profiles` ≥3, `same_workload` identity, and `placements` ≥3.

## Run

```bash
make test         # go test -race ./...
make demo-local   # builds, runs on :18408, curls /v1/demo → demo-output.json
make demo         # docker compose variant
make down
```

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-n-008:local`, apply with kubectl/Kind).

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info`
- `PUT /v1/profiles/{name}` — create/replace a profile (`runtime`, `constraints`)
- `POST /v1/workloads/{id}/schedule` — schedule workload across `profiles`
- `GET /v1/workloads/{id}/placements` — list placements
- `GET|POST /v1/demo` — live proof: `profiles`, `same_workload`, `placements`
- `GET /metrics`
