---
phase: 260711-x5m-setup-observability-comprehensive-lgtm-g
plan: 01
subsystem: observability
tags: [prometheus, loki, tempo, grafana, alloy, micrometer, otel, slo]

requires: []
provides:
  - Discrete LGTM + Alloy stack on kind (Prometheus metrics backend)
  - Micrometer /metrics + OTLP traces + JSON logs on all four services
  - Twelve provisioned Grafana dashboards with SLO/error-budget wiring
  - Multi-window burn alerts + error-budget runbook
affects: [kind-deploy, service-instrumentation, sre-demo]

tech-stack:
  added:
    - prom/prometheus:v2.54.1
    - grafana/loki:3.1.1
    - grafana/tempo:2.6.1
    - grafana/grafana:11.2.0
    - grafana/alloy:v1.3.1
    - kube-state-metrics:v2.13.0
    - micrometer-registry-prometheus
    - opentelemetry-sdk + OTLP exporter
    - logstash-logback-encoder
  patterns:
    - Shared Observability helper in common/
    - Alloy kubernetes SD via prometheus.io annotations
    - Recording rules sli:*/slo:* consumed by Grafana

key-files:
  created:
    - infra/k8s/observability/*.yaml
    - infra/k8s/observability/grafana/dashboards/01-12*.json
    - docs/runbooks/error-budget-burn.md
    - common/src/main/kotlin/com/marketplace/common/observability/Observability.kt
    - infra/compose-observability/*
  modified:
    - scripts/deploy-kind.sh
    - infra/k8s/kind-config.yaml
    - infra/docker-compose.yml
    - CLAUDE.md
    - infra/k8s/10-13 service Deployments
    - four service Application.kt + logback.xml

key-decisions:
  - "Prometheus (not Mimir) for kind memory limits"
  - "Shared common/ Observability helper to avoid duplicating Micrometer/OTel across 4 services"
  - "Grafana dashboards ConfigMap materialized by deploy-kind from JSON files"
  - "Manual OTel server spans instead of alpha ktor instrumentation artifact"

patterns-established:
  - "prometheus.io/scrape|port|path annotations on app pods"
  - "OTEL_EXPORTER_OTLP_ENDPOINT → Alloy :4318 → Tempo"
  - "Metric contract documented in 26-recording-rules.yaml header"

requirements-completed: [OBS-LGTM, OBS-INSTRUMENT, OBS-SLO-12]

duration: 5min
completed: 2026-07-12
---

# Phase 260711-x5m: Setup Observability (LGTM) Summary

**Kind-ready LGTM stack (Prometheus+Loki+Tempo+Grafana+Alloy) with Micrometer/OTLP on all four services and 12 SLO dashboards wired to recording rules and multi-window burn alerts.**

## Performance

- **Duration:** ~5 min
- **Started:** 2026-07-12T06:54:54Z
- **Completed:** 2026-07-12T06:59:52Z
- **Tasks:** 3/3
- **Files modified:** ~50 (across 3 atomic commits)

## Accomplishments

- Discrete single-replica LGTM + Alloy + kube-state-metrics under `infra/k8s/observability/` (no grafana/otel-lgtm all-in-one)
- All four Ktor services expose `/metrics`, export OTLP traces, emit JSON logs with `trace_id`/`span_id` MDC
- Exactly 12 Grafana dashboards provisioned; SLO board queries `sli:*` / `slo:*` recording rules; burn alerts link `docs/runbooks/error-budget-burn.md`

## Task Commits

1. **Task 1: LGTM k8s stack, Alloy, rules, deploy, runbook** - `c0fce23` (feat)
2. **Task 2: Instrument all four Ktor services** - `59df63a` (feat)
3. **Task 3: Provision 12 Grafana dashboards + SLO wiring** - `7dd85c6` (feat)

**Plan metadata:** skipped (orchestrator handles docs commit per quick-task constraints)

## Files Created/Modified

- `infra/k8s/observability/20-27-*.yaml` — Prometheus, Loki, Tempo, Grafana, Alloy, kube-state-metrics, rules, alerts
- `infra/k8s/observability/grafana/dashboards/*.json` — 12 dashboards
- `docs/runbooks/error-budget-burn.md` — burn-alert runbook
- `common/.../Observability.kt` — shared Micrometer + OTLP helper
- `scripts/deploy-kind.sh` — preload obs images, apply stack, create dashboard ConfigMap
- `CLAUDE.md` — env vars + observability section; removed "no observability stack"

## Decisions Made

- Prometheus for kind (documented in CLAUDE.md); Mimir-compatible PromQL API for dashboards
- Shared `common` observability module (acceptable plan deviation for DRY)
- `deploy-kind.sh` creates `grafana-dashboards` ConfigMap from JSON directory before apply

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing critical functionality] Shared Observability helper in common/**
- **Found during:** Task 2
- **Issue:** Duplicating Micrometer/OTel/log wiring across four services would drift
- **Fix:** `common/.../Observability.kt` + api deps; services call `installObservability`
- **Files modified:** `common/build.gradle.kts`, `Observability.kt`, four `Application.kt`
- **Committed in:** `59df63a`

**2. [Rule 3 - Blocking] Manual OTel spans instead of alpha Ktor instrumentation**
- **Found during:** Task 2 compile
- **Issue:** `opentelemetry-ktor-2.0` artifact unresolved / unstable with chosen versions
- **Fix:** W3C-propagating server span interceptor in shared helper
- **Verification:** `./gradlew :*:compileKotlin` succeeded
- **Committed in:** `59df63a`

**3. [Rule 3 - Blocking] Dashboard ConfigMap via deploy-kind**
- **Found during:** Task 3
- **Issue:** Twelve JSON files need ConfigMap mount without embedding giant YAML
- **Fix:** `kubectl create configmap grafana-dashboards --from-file=...` in `deploy-kind.sh`
- **Committed in:** `7dd85c6`

**Total deviations:** 3 auto-fixed (Rules 2–3)
**Impact on plan:** Cleaner instrumentation; no scope creep; must-haves intact

## Issues Encountered

- Micrometer 1.13 package rename (`io.micrometer.prometheusmetrics`) required import updates during Task 2 compile

## User Setup Required

None — kind/local demo only. After `./scripts/build-images.sh && ./scripts/deploy-kind.sh`, open http://localhost:3000 (anonymous Viewer).

## Known Stubs

None that block the plan goal. Infra health dashboards use cadvisor/kube-state proxies rather than dedicated Postgres/Redis/OpenSearch exporters (acceptable per plan for Task 1 kind footprint).

## Threat Flags

None beyond plan threat model (anonymous Grafana Viewer accepted for local demo; OTLP unauthenticated cluster-internal accepted).

## Next Phase Readiness

- Ready for kind smoke: deploy, hit `/metrics`, browse Grafana folder "C2C Marketplace"
- Optional follow-up: postgres/redis exporters for richer infra panels; Alertmanager UI

## Self-Check: PASSED

- All key artifacts present; commits `c0fce23`, `59df63a`, `7dd85c6` found; 12 dashboard JSONs present
