# ENG-E-014: Deep Kubernetes cluster-level platform integration

**Kind:** explicit | **Domain:** eng | **Stack:** go+k8s (stdlib only)

## Evidence from posting
Deep experience with Kubernetes... design platform services that integrate at that layer

## Rationale
Must integrate at scheduling/packaging/cluster APIs, not only kubectl.

## What this proves
- **CIJob** CRD schedules a **Job** child (`job_scheduled`)
- Terminal conditions **Complete** / **Failed** derived server-side from Job state
- Distinct from Phase 4 **ENG-N-007 ManagedWorkload** (finalizers are not the primary proof)

## Kind optional
`demo-local` (in-memory controller on `:18514`) is the phase gate. Applying `k8s/crd.yaml` to a Kind cluster is optional documentation, not required for acceptance.

## Acceptance demo
```bash
make test
make demo-local
```

Proof fields: `kind=CIJob`, `job_scheduled`, `complete`, `failed`, `complete_and_failed_paths`.

## Endpoints
- `GET /healthz`, `GET /readyz`, `GET /v1/info`
- `POST /v1/cijobs`, `GET /v1/cijobs/{id}`, `POST /v1/cijobs/{id}/outcome`
- `POST /v1/reconcile`
- `GET|POST /v1/demo`
- `GET /metrics`

## Dependencies
Go stdlib only — no controller-runtime.
