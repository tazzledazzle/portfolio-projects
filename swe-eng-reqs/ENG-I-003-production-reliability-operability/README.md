# ENG-I-003: Production reliability / operability

**Kind:** implicit | **Domain:** eng | **Stack:** go+compose

## Evidence from posting
post-mortems, quality, multi-DC scale, SLO gates nice-to-have

## Rationale
Platforms must be operable at CoreWeave scale.

## Acceptance demo
Operability kit with stored SLO definitions, an indexed local runbook library,
and a dashboard artifact for latency, traffic, errors, and saturation. This
slice owns operational readiness artifacts; it does not implement release gate
allow/deny decisions or telemetry span export.

## Run

```bash
make test
make demo-local
cat demo-output.json
```

The Docker-free proof runs on `127.0.0.1:18413`. Docker Compose remains
available:

```bash
make demo
make down
```

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-i-003:local`, apply with kubectl/Kind).

## Operability artifacts

- `runbooks/error-budget.md` — original, paraphrased response checklist with
  no copied book text and no secrets.
- `dashboards/golden-signals.json` — portable metric-name inventory for the
  four golden signals.

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info`
- `PUT /v1/slos/{id}`
- `GET /v1/golden-signals`
- `GET /v1/runbooks`
- `GET|POST /v1/demo`
- `GET /metrics`
