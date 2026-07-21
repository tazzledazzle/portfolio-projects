# ENG-E-015: LLM-integrated developer tooling interest or experience

**Kind:** explicit | **Domain:** eng | **Stack:** python+compose

## Evidence from posting
Experience with or strong technical interest in LLM-integrated developer tooling

## Rationale
Soft bar: real interest/experience required, not necessarily production LLM ownership.

## Acceptance demo
Offline-eval LLM workflow that summarizes failing pipelines from fixtures.

## Offline fixture LLM honesty

- `offline_fixture_llm=true`
- `simulator=true`
- `live_provider=false`
- Uses deterministic on-disk fixtures and Python stdlib only.
- Does not connect to OpenAI, Anthropic, or another live model provider and requires zero API keys.
- Scope is failure summarization only; agent loops, workflow orchestration, eval infrastructure, and ROI belong to other requirement slices.

## Run

```bash
make test
env -u OPENAI_API_KEY -u ANTHROPIC_API_KEY make demo-local
```

The local proof listens on `127.0.0.1:18815` and writes `demo-output.json`.

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-e-015:local`, apply with kubectl/Kind).

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info`
- `GET|POST /v1/demo`
- `POST /v1/summarize`
- `GET /metrics`
