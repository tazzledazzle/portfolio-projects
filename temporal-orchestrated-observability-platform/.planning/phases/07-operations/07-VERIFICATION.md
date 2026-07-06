# Phase 7 Verification — Operations & On-Call

**Verified:** 2026-05-20  
**Environment:** Local Docker Compose + host-run Kotlin worker  
**Requirements:** OPS-01 … OPS-03

## OPS-01 — Runbooks (`docs/OPERATIONS.md`)

Four failure modes documented with detection queries and executable remediation:

| Runbook | Sections |
|---------|----------|
| Stuck workflow | Temporal UI, Loki filter, terminate/reset, `SIMULATE_TOOL_FAILURE` |
| Missing traces | trace_id check, collector/tempo/jaeger health, OTLP env |
| Prometheus target down | `/targets`, port 9464, `host.docker.internal`, reload |
| Loki / Tempo disk pressure | `docker system df`, prune, retention, log truncate |

**Result:** PASS — all four modes have step-by-step procedures.

## OPS-02 — Alert rules

**File:** `deploy/prometheus/alerts.yml`  
**Loaded by:** `deploy/prometheus/prometheus.yml` → `rule_files`

| Alert | Intent |
|-------|--------|
| `HighActivityFailureRate` | Workflow error ratio > 25% / 5m |
| `AiWorkerTargetDown` | `up{job="ai-worker"} == 0` |
| `OtelExporterSendFailures` | Collector span export failures |
| `OtelCollectorTargetDown` | Collector metrics scrape down |

**Automated check:**

```bash
docker run --rm \
  -v "$PWD/deploy/prometheus:/etc/prometheus:ro" \
  prom/prometheus:v2.55.1 \
  promtool check rules /etc/prometheus/alerts.yml
```

**Result:** PASS (`promtool check rules` SUCCESS)  
**CI:** `.github/workflows/ci.yml` — Validate Prometheus alert rules step added.

## OPS-03 — Manual runbook validation

| Runbook | Procedure executed | Outcome | Notes |
|---------|-------------------|---------|-------|
| 1. Stuck workflow | `SIMULATE_TOOL_FAILURE=true` + `./gradlew :starter:run --args="agent test"` with worker running | PASS | Retries visible in Temporal UI; clears when env unset |
| 2. Missing traces | Stopped collector (`docker stop temporal-obs-platform-otel-collector-1`), ran `ping`, searched Jaeger/Tempo | PASS | No new traces while collector stopped; traces return after `docker start` + workflow |
| 3. Prometheus target down | Worker stopped; `curl localhost:9090/api/v1/targets` for `ai-worker` | PASS | Health `down` without worker; `up` after `./gradlew :worker:run` + metrics curl |
| 4. Loki / Tempo disk pressure | `docker system df`; documented prune/restart steps | PASS | Procedure validated (no forced OOM); prune command syntax verified |

**Validator:** GSD execute-phase 7 (automated + spot-check)  
**Date:** 2026-05-20

## Additional changes

- Prometheus `extra_hosts: host.docker.internal:host-gateway` for Linux scrape compatibility
- Prometheus volume mount: entire `deploy/prometheus/` directory

## Checklist

- [x] OPS-01 Complete OPERATIONS.md
- [x] OPS-02 Example alerts + promtool CI
- [x] OPS-03 Manual validation table with PASS per runbook
- [x] README portfolio polish

**Phase 7 complete. v1 milestone ready for `/gsd-transition` or portfolio publish.**
