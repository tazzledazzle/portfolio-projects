# Research Summary

**Project:** Temporal-Orchestrated Observability Platform  
**Date:** 2026-05-19

## Recommendation

Build a **Compose-based reference stack** in seven sequential GSD phases matching the architecture diagram. Prioritize **correlated telemetry** (Temporal IDs + OTel) over feature-rich AI—workflows use **stubs** until v2.

## Critical path

1. Foundation (Compose health)
2. Temporal worker connectivity
3. Sample workflows (proves orchestration value)
4. OTel instrumentation (proves Kotlin + OTel layer)
5. Jaeger + Prometheus (proves export path)
6. LGTM Grafana (proves operator story)
7. Operations (proves on-call readiness)

## Top risks

1. OTel + Temporal Kotlin interceptor gaps → manual spans
2. Resource exhaustion on laptop → Compose memory limits
3. Trace backend duality → clear phase cutover in docs

## Stack headline

**Kotlin · Temporal · OpenTelemetry · Prometheus · Jaeger/Tempo · Loki · Grafana**

## Ready for planning

Requirements and roadmap trace 22 v1 requirements across 7 phases. Proceed with **Phase 1 execution** after stakeholder review of `docs/ARCHITECTURE.md` and ADRs.
