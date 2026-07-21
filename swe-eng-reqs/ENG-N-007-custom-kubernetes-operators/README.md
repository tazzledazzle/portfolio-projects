# ENG-N-007: Custom Kubernetes operators

**Kind:** nice | **Domain:** eng | **Stack:** go+k8s

## Evidence from posting
custom Kubernetes operators

## Rationale
Extends deep K8s into operator authorship.

## Acceptance demo
`ManagedWorkload` operator proof with an in-memory reconcile loop, `Ready`
conditions, and deletion finalizers. The local demo creates a workload,
reconciles it, marks it deleting, and finalizes cleanup.

## Run

```bash
make test
make demo-local
cat demo-output.json
```

`demo-local` is the required gate. It builds the service and runs the live
proof on `127.0.0.1:18407` without Docker, Kind, or a kubeconfig.

Docker Compose remains available:

```bash
make demo
make down
```

## Optional Kind CRD proof

Kind is optional and is not a phase or `demo-local` gate. When a Kind cluster
is already available, apply the CRD separately:

```bash
kubectl apply -f k8s/crd.yaml
kubectl get crd managedworkloads.devex.coreweave.example
```

The CRD defines `ManagedWorkload` with `spec.replicas`, `spec.image`, and
list-map `status.conditions`. `k8s/deploy.yaml` deploys the demo service after
building the `eng-n-007:local` image.

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info`
- `POST /v1/workloads`
- `POST /v1/reconcile`
- `DELETE /v1/workloads/{id}`
- `POST /v1/workloads/{id}/finalize`
- `GET|POST /v1/demo`
- `GET /metrics`
