# C2C Marketplace K8s

Portfolio C2C marketplace: four Ktor/Kotlin microservices (listings, search, messaging, payments) on kind, with Postgres, Redis, OpenSearch, and Redpanda.

## Current focus

Observability — LGTM (Loki, Grafana, Tempo, Mimir/Prometheus) stack with comprehensive signal collection and SLO/SLA dashboards.

## Known deliberate omissions (baseline)

No auth/authz, no API gateway, logs-to-stdout only until observability is added.
