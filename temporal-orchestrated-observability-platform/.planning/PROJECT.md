# Temporal-Orchestrated Observability Platform

## What This Is

A **reference platform** that runs sample **AI workflows** on **Temporal**, instruments execution with **Kotlin + OpenTelemetry**, and surfaces **metrics, traces, and logs** in **Grafana (LGTM)** so operators get **end-to-end visibility** from workflow start through LLM/retrieval activities to on-call response.

## Core Value

**An operator can start an AI workflow, watch it succeed or fail in Temporal, and within Grafana trace the same run across metrics, traces, and logs using shared workflow identifiers.**

## Current State (v1.0 — shipped 2026-05-20)

- **Worker:** Kotlin/JVM 21 Gradle monorepo (~1,450 LOC) — `:workflows`, `:worker`, `:starter`
- **Workflows:** PingWorkflow, RagQa, AgentTools, BatchEval with WireMock LLM stub
- **Stack:** Docker Compose — Temporal, Grafana, Loki, Tempo, Prometheus, Jaeger, OTel Collector, Promtail
- **Telemetry:** OTel traces/metrics/logs correlated by `workflow_id`; Prometheus `:9464/metrics`
- **Operations:** `docs/OPERATIONS.md` runbooks, `deploy/prometheus/alerts.yml`, smoke scripts
- **Evidence:** `.planning/phases/*/VERIFICATION.md` · archived requirements in `.planning/milestones/v1.0-REQUIREMENTS.md`

## Requirements

### Validated (v1.0)

- ✓ **Foundation (Phase 1)** — Gradle modules, Compose LGTM+Temporal stack, smoke script, CI
- ✓ **Temporal (Phase 2)** — PingWorkflow, worker, starter CLI
- ✓ **AI workflows (Phase 3)** — RagQa, AgentTools, BatchEval + LLM stub
- ✓ **Instrumentation (Phase 4)** — OTel traces, metrics, JSON logs, OkHttp propagation
- ✓ **Backends (Phase 5)** — Jaeger + Prometheus scrape, collector validate
- ✓ **LGTM (Phase 6)** — Loki, Tempo, Grafana dashboards, log↔trace correlation
- ✓ **Operations (Phase 7)** — Runbooks, Prometheus alerts, manual validation

### Active (next milestone — not yet defined)

Run `/gsd-new-milestone` to define v1.1+ requirements. v2 candidates from prior scoping:

- **K8S-01** — Helm chart or Kustomize for Temporal + observability stack
- **K8S-02** — Temporal mTLS between worker and server
- **AI-01** — Real OpenAI/Anthropic integration behind feature flag
- **AI-02** — LangGraph activity embedding inside Temporal activities

### Out of Scope

| Feature | Reason |
|---------|--------|
| Multi-region Temporal | Portfolio scope |
| Custom embedding training | Not observability focus |
| SaaS auth/billing | Demo platform |
| PagerDuty integration | Example alerts only (v1) |
| Production Temporal Cloud | Self-hosted Compose for v1 |

## Next Milestone Goals

Prioritize from v2 candidates above or add new scope via `/gsd-new-milestone`:

1. **Portfolio polish** — Grafana screenshot, README demo flow
2. **Kubernetes stretch** — Helm/Kustomize deployment path
3. **Real LLM integration** — feature-flagged provider behind existing stub interface

## Context

Greenfield project from architecture diagram **"5. Temporal-Orchestrated Observability Platform"** (2026-05-19).  
Design: `docs/ARCHITECTURE.md` · ADRs: `docs/adr/`

## Constraints

- **Kotlin/JVM 21** workers with Gradle
- **Temporal** 1.25+ SDK
- **OpenTelemetry** Java SDK; OTLP + Prometheus exporters
- **Docker Compose** for all infrastructure in v1
- **Portfolio credibility** — reproducible smoke, honest docs

## Key Decisions

| Decision | Rationale | ADR | Outcome |
|----------|-----------|-----|---------|
| Temporal for orchestration | Durable AI steps, UI, retries | ADR-001 | ✓ Good — all workflows durable with visible history |
| Kotlin + OTel workers | Matches diagram; LGTM-native telemetry | ADR-002 | ✓ Good — full trace/metric/log correlation |
| Jaeger dev → Tempo LGTM | Fast debug + Grafana correlation | ADR-003 | ✓ Good — dual-export supports both paths |
| workflow_id on traces/logs, not metric labels | Avoid cardinality explosion | ADR-004 | ✓ Good — bounded Prometheus labels |
| Compose-first platform | Reproducible local demo | ADR-005 | ✓ Good — single-command stack bring-up |

---
*Last updated: 2026-06-22 after v1.0 milestone*
