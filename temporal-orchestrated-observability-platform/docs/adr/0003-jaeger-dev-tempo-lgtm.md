# ADR-003: Jaeger for Dev Traces, Tempo for LGTM

## Status
Accepted

## Context
The architecture diagram shows **Jaeger** receiving traces from OpenTelemetry and the **LGTM stack** using **Tempo** for trace storage in Grafana. Running both permanently adds complexity; running only Jaeger breaks Grafana trace correlation.

## Decision
- **Phase 5 (dev loop):** Export OTLP traces to **Jaeger** for fast local debugging.
- **Phase 6 (LGTM):** Add **Grafana Tempo** as the Grafana-native trace store; configure OTel Collector to **dual-export** or **switch** primary backend to Tempo.
- **Jaeger** remains optional in Compose after Phase 6 (profile `jaeger`) for developers who prefer Jaeger UI.

## Alternatives Considered

- **Tempo only from day one** — Simpler long-term; slower initial feedback without Jaeger's mature UI.
- **Jaeger only** — Avoids Tempo; weaker "LGTM" story and Grafana trace linking.
- **Elastic APM** — Heavier stack; off-portfolio.

## Consequences

### Positive
- Matches diagram layers explicitly in build phases
- Grafana Explore can correlate logs (Loki) ↔ traces (Tempo)
- Jaeger accelerates early instrumentation debugging

### Negative
- Temporary dual backends during migration
- Collector config must be validated (OBS risk)

## Trade-offs
Short-term dual export is accepted for **faster instrumentation iteration** without sacrificing the LGTM end state.
