# ADR-005: Docker Compose for Local LGTM Stack

## Status
Accepted

## Context
Portfolio deliverable must be reproducible on a laptop without cloud accounts. Kubernetes adds friction for first contributors.

## Decision
Provide **`deploy/docker-compose.yml`** (and profiles) that runs:

- Temporal Server (+ UI)
- OpenTelemetry Collector
- Prometheus, Loki, Promtail, Tempo, Grafana
- Jaeger (profile `jaeger` or Phase 5 default)
- Optional LLM stub (WireMock or similar)

Kubernetes manifests are **out of v1 scope** but noted in Phase 7 docs as a stretch path.

## Alternatives Considered

- **Grafana Cloud only** — Simpler ops; requires API keys and breaks offline demo.
- **minikube + Helm** — Closer to prod; slower iteration for observability tuning.
- **Podman compose** — Compatible; document as alternative, standardize on Docker Compose v2 syntax.

## Consequences

### Positive
- One-command platform for demos and CI smoke
- Matches LGTM learning path in Grafana docs

### Negative
- Resource-heavy (~8 GB RAM recommended)
- Not production HA

## Trade-offs
Compose-first accelerates **observability integration**; K8s deferred to v2.
