# Roadmap: Temporal-Orchestrated Observability Platform

## Overview

Build bottom-up along the architecture diagram: **local platform → Temporal → AI workflows → Kotlin/OTel → Jaeger/Prometheus → LGTM/Grafana → operations**.

## Phases

| # | Phase | Goal |
|---|-------|------|
| 1 | [Foundation](phases/01-foundation/) | Reproducible repo + Compose + smoke ✓ |
| 2 | [Temporal orchestration](phases/02-temporal/) | Server + Kotlin worker shell ✓ |
| 3 | [AI workflows](phases/03-ai-workflows/) | Three sample Temporal workflows ✓ |
| 4 | [Instrumentation](phases/04-instrumentation/) | Kotlin + OpenTelemetry |
| 5 | [Trace & metrics backends](phases/05-backends/) | Prometheus + Jaeger OTLP |
| 6 | [LGTM & Grafana](phases/06-lgtm/) | Loki, Tempo, dashboards |
| 7 | [Operations](phases/07-operations/) | Runbooks, alerts, verification |

## Phase details

### Phase 1: Foundation

**Goal:** Any contributor can clone, `docker compose up`, and run `./scripts/smoke.sh` successfully.  
**Depends on:** Nothing  
**Requirements:** FOUND-01, FOUND-02, FOUND-03  
**Success criteria:**

1. Compose starts Temporal, Grafana, Prometheus, Loki, Tempo without manual fixes.
2. Smoke script validates HTTP health for key services.
3. CI passes Gradle build on empty worker module.

**Plans:** [01-01-PLAN.md](phases/01-foundation/01-01-PLAN.md)

---

### Phase 2: Temporal orchestration

**Goal:** Kotlin worker polls Temporal and executes a no-op "ping" workflow end-to-end.  
**Depends on:** Phase 1  
**Requirements:** TEMP-01, TEMP-02, TEMP-03  
**Success criteria:**

1. Worker logs show successful poll on `ai-workflows`.
2. Temporal UI displays completed `PingWorkflow` execution.
3. Documented shutdown procedure leaves no orphaned tasks.

**Plans:** [02-01-PLAN.md](phases/02-temporal/02-01-PLAN.md)

---

### Phase 3: AI workflows

**Goal:** Three reference AI workflows run with stubbed external dependencies.  
**Depends on:** Phase 2  
**Requirements:** WF-01 … WF-04  
**Success criteria:**

1. RAG, agent, and batch workflows each complete at least once via starter CLI.
2. Retry policy demonstrated on agent workflow (injected failure).
3. Batch workflow produces aggregate result visible in workflow return value.

**Plans:** [03-01-PLAN.md](phases/03-ai-workflows/03-01-PLAN.md)

---

### Phase 4: Instrumentation (Kotlin + OTel)

**Goal:** Worker emits correlated traces, metrics, and logs for all activities.  
**Depends on:** Phase 3  
**Requirements:** OTEL-01 … OTEL-04  
**Success criteria:**

1. Sample run produces spans with Temporal identifiers in Jaeger (once Phase 5 wired) or console exporter.
2. Prometheus metrics endpoint exposes bounded-label histograms.
3. JSON logs include `trace_id` when activities run.

**Plans:**

1. [04-01-PLAN.md](phases/04-instrumentation/04-01-PLAN.md) — OTel SDK, Temporal interceptors, Prometheus metrics (OTEL-01, OTEL-02)
2. [04-02-PLAN.md](phases/04-instrumentation/04-02-PLAN.md) — JSON logs, OkHttp tracing, smoke script (OTEL-03, OTEL-04)

---

### Phase 5: Trace & metrics backends

**Goal:** OTLP traces in Jaeger; Prometheus scrapes worker.  
**Depends on:** Phase 4  
**Requirements:** BACK-01, BACK-02, BACK-03  
**Success criteria:**

1. Jaeger UI shows multi-span trace for one workflow run.
2. Prometheus targets page shows worker `UP`.
3. CI validates OTel Collector configuration.

**Plans:** [05-01-PLAN.md](phases/05-backends/05-01-PLAN.md)

---

### Phase 6: LGTM & Grafana

**Goal:** Unified Grafana experience with logs, traces, metrics.  
**Depends on:** Phase 5  
**Requirements:** LGTM-01 … LGTM-04  
**Success criteria:**

1. Loki query by `workflow_id` returns activity logs.
2. Grafana Explore opens Tempo trace from log derived field.
3. Three provisioned dashboards render non-empty panels after sample load.

**Plans:** [06-01-PLAN.md](phases/06-lgtm/06-01-PLAN.md)

---

### Phase 7: Operations

**Goal:** Operators can detect and respond to common failures.  
**Depends on:** Phase 6  
**Requirements:** OPS-01 … OPS-03  
**Success criteria:**

1. OPERATIONS.md complete with four failure modes.
2. Example Prometheus alert rules committed and lint-valid.
3. Verification doc records manual runbook execution dates.

**Plans:** [07-01-PLAN.md](phases/07-operations/07-01-PLAN.md)

---

## Progress

| Phase | Plans | Status | Completed |
|-------|-------|--------|-----------|
| 1. Foundation | 1/1 | **Complete** | 2026-05-20 |
| 2. Temporal | 1/1 | **Complete** | 2026-05-20 |
| 3. AI workflows | 1/1 | **Complete** | 2026-05-20 |
| 4. Instrumentation | 0/1 | Not started | — |
| 5. Backends | 0/1 | Not started | — |
| 6. LGTM | 0/1 | Not started | — |
| 7. Operations | 0/1 | Not started | — |

---
*Roadmap created: 2026-05-19*
