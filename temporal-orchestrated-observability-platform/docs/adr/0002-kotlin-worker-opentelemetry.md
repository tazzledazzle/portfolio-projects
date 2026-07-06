# ADR-002: Kotlin Workers with OpenTelemetry SDK

## Status
Accepted

## Context
Workers execute AI activities and are the natural instrumentation boundary. We need vendor-neutral traces, metrics, and logs that integrate with Grafana LGTM. The diagram specifies **Kotlin + OTel**.

## Decision
Implement Temporal workers in **Kotlin (JVM 21)** using the **OpenTelemetry Java SDK** (API + SDK + OTLP exporter + Prometheus exporter). Use Temporal OTel interceptors when available; supplement with manual spans for LLM HTTP clients.

## Alternatives Considered

- **Java workers** — Equivalent OTel support; Kotlin chosen for coroutine ergonomics and portfolio differentiation.
- **Python workers** — Faster AI prototyping; weaker alignment with "Kotlin + OTel" diagram and JVM Temporal performance story.
- **Datadog/New Relic agents** — Faster dashboard setup; violates vendor-neutral LGTM goal.

## Consequences

### Positive
- Single process exports traces (OTLP) and metrics (Prometheus scrape)
- Aligns with Grafana/OpenTelemetry ecosystem
- Strong typing for workflow/activity context propagation

### Negative
- JVM memory footprint vs Python
- OTel + Temporal interceptor documentation sparse for Kotlin

## Trade-offs
Kotlin + manual instrumentation effort is preferred over proprietary APM for **portable, reproducible observability**.
