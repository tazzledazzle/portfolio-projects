# ENG-I-005: Observability for platforms

**Kind:** implicit | **Domain:** eng | **Stack:** go+compose

## Evidence from posting
scale + quality + SLO gates + postmortems

## Rationale
Cannot operate 40+ DC platforms without observability.

## Acceptance demo
OTel-inspired stdlib simulator exporting in-process spans and metrics and
evaluating fixed threshold alert rules.

> **Honesty boundary:** this is an `otel_inspired` simulator with
> `instrumentation_model: otel-inspired` and `collector: none`. It does not use
> the real OpenTelemetry SDK and does not connect to an OTel collector, Tempo,
> or Grafana.

## Run

```bash
make test
make demo-local
cat demo-output.json
```

The Docker-free live proof runs on `127.0.0.1:18415`. Docker Compose remains
available:

```bash
make demo
make down
```

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-i-005:local`, apply with kubectl/Kind).

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info` — includes `otel_inspired`, `simulator`,
  `instrumentation_model`, and `collector`
- `POST /v1/trace`
- `GET /v1/traces`
- `POST /v1/alerts/evaluate` — fixed fixture rules only
- `GET|POST /v1/demo`
- `GET /metrics`
