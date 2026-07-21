# ENG-I-007: Developer interface quality (API/CLI UX)

**Kind:** implicit | **Domain:** eng | **Stack:** go+compose

## Evidence from posting
IDP + API systems design + translation

## Rationale
DevEx success is interface quality for engineers. This slice owns CLI UX
(`status --json`, clear errors, golden snapshots). It does **not** own server
OpenAPI / rate limits (ENG-E-021) or IDP catalog nouns (ENG-E-005).

## Acceptance demo
CLI with `--json`, clear errors, and golden snapshot tests. Live `/v1/demo`
proof fields: `cli_json`, `clear_errors`, `golden_match` on port **18607**.

## CLI

```bash
go run . status --json
# or after build:
./eng-i-007 status --json
```

Unknown commands and missing `--json` return clear stderr messages (no panic,
no secret leakage).

## Run

```bash
make test
make demo-local
make demo
make down
```

Kubernetes manifests: `k8s/deploy.yaml` (build image `eng-i-007:local`, apply with kubectl/Kind).

## Endpoints
- `GET /healthz`
- `GET /readyz`
- `GET /v1/info`
- `GET|POST /v1/demo` — exercises CLI and returns proof
- `GET /metrics`
