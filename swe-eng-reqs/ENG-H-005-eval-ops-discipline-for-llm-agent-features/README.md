# ENG-H-005: Eval/ops discipline for LLM/agent features

**Kind:** hidden | **Domain:** eng | **Stack:** python+compose

## Evidence from posting
LLM/MCP/AI workflows + DevEx value nice-to-have

## Rationale
Without evals, AI DevEx claims are indefensible.

## Acceptance demo
Offline/online eval harness with known failure fixtures.

## SIMULATOR / offline evals

The harness runs golden and known-bad fixtures entirely offline. Its
`online-sim` mode injects a local deterministic callable to exercise the same
operational boundary without network access, API keys, or a live model.
Known-bad prompt-injection output must score as failed and caught.

## Run

```bash
make test
make demo-local  # listens on :18905, no API keys or Docker required
```

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-h-005:local`, apply with kubectl/Kind).

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info`
- `GET|POST /v1/demo`
- `POST /v1/evals/run` (`{"mode":"offline"}` or `online-sim`)
- `GET /metrics`
