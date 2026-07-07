# Roadmap: Temporal-Orchestrated Observability Platform

## Milestones

- ✅ **v1.0 Reference Platform** — Phases 1–7 (shipped 2026-05-20)

## Overview

Build bottom-up along the architecture diagram: **local platform → Temporal → AI workflows → Kotlin/OTel → Jaeger/Prometheus → LGTM/Grafana → operations**.

## Phases

<details>
<summary>✅ v1.0 Reference Platform (Phases 1–7) — SHIPPED 2026-05-20</summary>

| # | Phase | Goal | Completed |
|---|-------|------|-----------|
| 1 | [Foundation](phases/01-foundation/) | Reproducible repo + Compose + smoke | 2026-05-20 |
| 2 | [Temporal orchestration](phases/02-temporal/) | Server + Kotlin worker shell | 2026-05-20 |
| 3 | [AI workflows](phases/03-ai-workflows/) | Three sample Temporal workflows | 2026-05-20 |
| 4 | [Instrumentation](phases/04-instrumentation/) | Kotlin + OpenTelemetry | 2026-05-20 |
| 5 | [Trace & metrics backends](phases/05-backends/) | Prometheus + Jaeger OTLP | 2026-05-20 |
| 6 | [LGTM & Grafana](phases/06-lgtm/) | Loki, Tempo, dashboards | 2026-05-20 |
| 7 | [Operations](phases/07-operations/) | Runbooks, alerts, verification | 2026-05-20 |

Full phase details: [milestones/v1.0-ROADMAP.md](milestones/v1.0-ROADMAP.md)

</details>

## Progress

| Phase | Milestone | Plans | Status | Completed |
|-------|-----------|-------|--------|-----------|
| 1. Foundation | v1.0 | 1/1 | Complete | 2026-05-20 |
| 2. Temporal | v1.0 | 1/1 | Complete | 2026-05-20 |
| 3. AI workflows | v1.0 | 1/1 | Complete | 2026-05-20 |
| 4. Instrumentation | v1.0 | 2/2 | Complete | 2026-05-20 |
| 5. Backends | v1.0 | 1/1 | Complete | 2026-05-20 |
| 6. LGTM | v1.0 | 1/1 | Complete | 2026-05-20 |
| 7. Operations | v1.0 | 1/1 | Complete | 2026-05-20 |

## Next Milestone

Use `/gsd-new-milestone` to define v1.1+ scope. See `.planning/PROJECT.md` for v2 candidates.

---
*Last updated: 2026-06-22 after v1.0 milestone archive*
