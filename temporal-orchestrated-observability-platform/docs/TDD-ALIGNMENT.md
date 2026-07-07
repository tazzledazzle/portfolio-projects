# TDD Alignment — v1 vs Design Document

**Source:** `AI_Workflow_Observability_TDD.docx` (May 2026)  
**v1 baseline:** 2026-05-20 — all 22 portfolio requirements complete  
**v2 target:** Close gaps to production-grade TDD (see [ROADMAP-v2.md](../.planning/ROADMAP-v2.md))

## Summary

| Dimension | v1 portfolio | TDD production target |
|-----------|--------------|------------------------|
| Deployment | Docker Compose | Kubernetes + Helm |
| Trace entry point | Worker activities | Agent → workflow → activity |
| Span schema | Partial (4 Temporal fields) | Full §5 schema incl. tokens |
| Dashboards | 3 | 4 |
| Alerting | Example rules | SLO + error budget + Alertmanager |
| Security | Demo defaults | TLS, SSO, PII scrub |

## Gap register (feeds v2 phases)

| ID | TDD reference | v1 state | v2 phase |
|----|---------------|----------|----------|
| G-01 | §4.2 — client W3C propagation | Missing | 8 |
| G-02 | §4.3.1 — header context on activity | Partial (no headers) | 8 |
| G-03 | §5 — `activity.attempt`, `worker.id`, `environment` | Missing | 8 |
| G-04 | §5 — `model.name`, `token.input/output` | Missing | 8, 13 |
| G-05 | §4.3.1 — `activity.retry_count`, `ai.token.total` | Missing | 8 |
| G-06 | §4.3.1 — span name `temporal.activity.*` | Uses `activity.*` | 8 |
| G-07 | §4.3.2 — collector processors | `batch` only | 9 |
| G-08 | §4.3.2 — logs via collector → Loki | Promtail file | 9 |
| G-09 | §4.3.3 — Tempo 30d / object storage | 48h local | 9 |
| G-10 | §4.3.4 — recording rules, Alertmanager | Missing | 11 |
| G-11 | §4.3.6 — Worker Health dashboard | Missing | 10 |
| G-12 | §4.3.6 — Trace Explorer dashboard | Missing | 10 |
| G-13 | §4.3.6 — Oncall Runbook Links dashboard | Missing | 10 |
| G-14 | §7.2 — SLO alerts (99.5%, p99, burn) | Partial | 11 |
| G-15 | §8 — TLS/mTLS OTLP | `insecure: true` | 12 |
| G-16 | §8 — Grafana OAuth/RBAC | admin/admin | 12 |
| G-17 | §8 — PII scrub in collector | Missing | 12 |
| G-18 | §7.1 — Helm / K8s IaC | Compose only | 9 |
| G-19 | OQ-3 — Agent instrumentation | Worker only | 8, 13 |
