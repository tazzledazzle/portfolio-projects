# Research — Stack

**Researched:** 2026-05-19

## Core runtime

| Layer | Technology | Version hint |
|-------|------------|--------------|
| Language | Kotlin | 2.0+ |
| JVM | Temurin | 21 |
| Build | Gradle (Kotlin DSL) | 8.x |
| Orchestration | Temporal Java/Kotlin SDK | 1.25+ |
| Telemetry | OpenTelemetry Java | 1.40+ |

## Infrastructure (Compose)

| Service | Image family | Notes |
|---------|--------------|-------|
| Temporal | `temporalio/auto-setup` | Dev single-node |
| OTel Collector | `otel/opentelemetry-collector-contrib` | OTLP + prometheus exporter |
| Jaeger | `jaegertracing/all-in-one` | Phase 5 |
| Prometheus | `prom/prometheus` | v2.x |
| Loki | `grafana/loki` | 3.x single binary |
| Promtail | `grafana/promtail` | Scrape worker logs |
| Tempo | `grafana/tempo` | OTLP enabled |
| Grafana | `grafana/grafana` | Provision datasources + dashboards |
| LLM stub | WireMock or `ghcr.io` minimal httpbin | Deterministic CI |

## Gradle dependencies (worker)

- `io.temporal:temporal-sdk`
- `io.opentelemetry:opentelemetry-api`, `sdk`, `exporter-otlp`, `exporter-prometheus`
- `io.opentelemetry.instrumentation:opentelemetry-instrumentation-annotations`
- `com.squareup.okhttp3:okhttp` (LLM HTTP with OTel okhttp instrumentation)
- `ch.qos.logback:logback-classic` + JSON encoder (logstash or logback-json)

## CI

- GitHub Actions: `./gradlew test build`, `docker compose config`, `otelcol validate` on collector config

## Not in v1

- Kubernetes / Helm
- Temporal Cloud
- Mimir (use Prometheus; note migration in docs)
