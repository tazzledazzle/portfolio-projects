# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

**Bootstrap (one-time, requires system Gradle):**
```bash
gradle wrapper --gradle-version 8.7
```

**Inner loop (docker-compose infra + Gradle run):**
```bash
cd infra && docker compose up -d   # postgres, redis, opensearch, redpanda
cd ..
./gradlew :listings-service:run    # port 8081
./gradlew :search-service:run      # port 8082
./gradlew :messaging-service:run   # port 8083
./gradlew :payments-service:run    # port 8084
```

Optional local LGTM (does not start with default compose):
```bash
cd infra && docker compose --profile observability up -d
# Grafana http://localhost:3000 — set OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
```

**Full k8s (kind):**
```bash
./scripts/build-images.sh    # gradle installDist + docker build for all 4 services
./scripts/deploy-kind.sh     # kind create cluster, load images, kubectl apply, Gloo+Flagger
kubectl -n c2c get pods -w
# Grafana: http://localhost:3000 (anonymous Viewer)
# Listings via Gloo: curl -H 'Host: listings.local' http://localhost:8081/healthz
kind delete cluster --name c2c-marketplace   # tear down
```

**Progressive delivery (listings canary):** Flagger + Gloo Edge (default). Metrics from existing LGTM Prometheus (`http://prometheus.c2c:9090`). Design: `docs/plans/2026-07-20-flagger-gloo-progressive-delivery-design.md`.
```bash
./scripts/canary-listings.sh                 # build tagged image, trigger canary, watch
PROGRESSIVE_PROVIDER=istio ./scripts/deploy-kind.sh   # Istio fallback instead of Gloo
```

## Tests

```bash
./gradlew :payments-service:test    # pure state machine unit tests — no Docker needed
./gradlew :listings-service:test    # Testcontainers (real Postgres) — requires Docker
./gradlew test                      # all modules
```

Run a single test class:
```bash
./gradlew :payments-service:test --tests "com.marketplace.payments.EscrowStateMachineTest"
```

## Architecture

Four Ktor/Kotlin services, each with its own Postgres schema (no shared DB). `common/` is a shared Gradle module holding `Events.kt` (Kotlin data classes annotated with `kotlinx.serialization`) that both Kafka producers and consumers depend on — this is the schema contract between services. Observability helpers live under `common/.../observability` (Micrometer `/metrics`, OTLP, JSON logging).

**Service responsibilities and key classes:**

| Service | Port | Key classes | Infra |
|---|---|---|---|
| listings-service | 8081 | `ListingRoutes`, `ListingRepository` (Exposed ORM), `EventPublisher` | Postgres, Kafka producer |
| search-service | 8082 | `ListingIndexer` (Kafka consumer), `OpenSearchClient`, `SearchRoutes` | Kafka consumer, OpenSearch |
| messaging-service | 8083 | `ChatWebSocket`, `ConnectionRegistry` | Postgres, Redis pub/sub |
| payments-service | 8084 | `EscrowStateMachine`, `OrderRepository`, `PaymentsRoutes` | Postgres, Kafka producer |

**Event flow:** `listings-service` writes a row to Postgres and publishes `listing.created` to Kafka. `search-service` consumes that event and indexes into OpenSearch (eventually consistent — acceptable for browse/search, not for payments). OpenSearch is the only store for `search-service`; replaying the Kafka topic from offset 0 fully rebuilds the index.

**Messaging pod routing:** `ConnectionRegistry` solves cross-pod WebSocket delivery using two Redis structures: a `presence` hash (`userId → podId`) and per-pod pub/sub channels (`pod:<podId>`). Each pod subscribes to its own channel at startup and fans messages out to locally-held sockets. The `podId` is a random UUID generated at startup.

**Payments consistency boundary:** `EscrowStateMachine` is intentionally pure (no I/O) — it only decides whether a `(EscrowStatus, EscrowEvent)` transition is legal, throwing `IllegalEscrowTransitionException` on illegal transitions. `OrderRepository` handles the two-write transaction (order row + escrow_hold row) that must commit atomically. Illegal transitions return HTTP 409, not 500.

**Escrow states:** `HELD → RELEASED` (confirm delivery or protection window elapsed) or `HELD → REFUNDED` (buyer dispute). Any transition from RELEASED or REFUNDED is illegal.

**Observability (kind):** Discrete LGTM stack under `infra/k8s/observability/` — **Prometheus** (metrics; preferred over Mimir for kind memory limits; PromQL-compatible), Loki (logs), Tempo (traces), Grafana (dashboards, anonymous Viewer on `:3000`), Grafana Alloy (DaemonSet: pod scrapes via `prometheus.io/*` annotations, stdout→Loki, OTLP→Tempo). SLOs: 99.9% availability (non-5xx), 99% latency &lt; 500ms; burn alerts + runbook at `docs/runbooks/error-budget-burn.md`. Grafana anonymous Viewer is for local kind demos only — do not expose publicly.

**Progressive delivery (kind):** Gloo Edge gateway-proxy owns kind NodePort `30081` → host `:8081` for listings. Flagger Canary on `listings-service` shifts weight via Gloo `RouteTable`; analysis queries Micrometer MetricTemplates against `prometheus.c2c:9090` (Datadog stand-in). Manifests under `infra/k8s/progressive/`. Search/messaging/payments remain direct NodePorts.

## Environment Variables

All services read config from env vars with localhost defaults for inner-loop dev:

| Var | Default | Used by |
|---|---|---|
| `DB_URL` | `jdbc:postgresql://localhost:5432/marketplace` | listings, messaging, payments |
| `DB_USER` / `DB_PASSWORD` | `marketplace` / `marketplace` | listings, messaging, payments |
| `KAFKA_BOOTSTRAP_SERVERS` | `localhost:9092` | listings (producer), search (consumer), payments (producer) |
| `REDIS_URL` | `redis://localhost:6379` | messaging |
| `OPENSEARCH_URL` | `http://localhost:9200` | search |
| `PORT` | 8081–8084 (per service) | all |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `http://localhost:4318` (local) / `http://alloy:4318` (k8s) | all (OTLP HTTP → Alloy → Tempo) |
| `OTEL_SERVICE_NAME` | per-service name | all |

## Known Deliberate Omissions

No auth/authz — `userId` is trusted from the request body. No full API gateway for all services (Gloo fronts listings only for the Flagger canary pilot). Shipping (`/orders/{id}/confirm-delivery` is the only payment transition exposed). Cross-service integration tests are not built (would need a docker-compose test harness or Kafka contract tests). Observability and progressive delivery are kind/demo-oriented (single-replica, short retention, anonymous Grafana Viewer) — not a production HA stack.
