# ENG-N-014: Agentic systems with DevEx value

**Kind:** nice | **Domain:** eng | **Stack:** python+compose

## Evidence from posting
Direct experience building agentic systems... incorporating these capabilities into DevEx platform features with consistent value-add

## Rationale
Preference for shipped agent systems with measured DevEx value.

## Acceptance demo
Agent feature with offline eval and ROI narrative (time saved / MTTR).

## Measurement honesty

The demo computes `time_saved_minutes` and `mttr_improvement_pct` from an
on-disk synthetic baseline and a deterministic assisted path. It always reports
`baseline_source=fixture` and `fabricated_prod=false`; these values are a
testable value hypothesis, not real production telemetry.

## Run

```bash
make test
make demo-local
```

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-n-014:local`, apply with kubectl/Kind).

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info`
- `GET|POST /v1/demo`
- `GET /metrics`
