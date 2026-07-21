# C2C marketplace (mock/simulated)

A runnable, simplified simulation of a mid-scale mobile-first C2C marketplace — four Kotlin/Ktor services (listings, search, messaging, payments) backed by Postgres, Redis, OpenSearch, and Kafka (via Redpanda), deployable to a local kind cluster.

Full design rationale, data flow diagrams, and known simplifications: [`docs/TDD.md`](docs/TDD.md).

## Prerequisites

- JDK 17+
- Docker (with enough memory allocated — OpenSearch alone wants ~1GB; Gloo+Flagger+LGTM need more)
- [kind](https://kind.sigs.k8s.io/) and `kubectl`, for the full k8s path
- [Helm 3](https://helm.sh/) (Gloo Edge + Flagger install in `deploy-kind.sh`)
- This repo has **no Gradle wrapper jar checked in** (binary files don't belong in a hand-written scaffold). Generate one once, locally:
  ```bash
  gradle wrapper --gradle-version 8.7
  ```
  (requires a system Gradle install just for this one bootstrap step — after that, `./gradlew` is self-contained.)

## Option A: fast inner loop (docker-compose, no k8s)

Best for actually writing and iterating on code.

```bash
cd infra
docker compose up -d          # postgres, redis, opensearch, redpanda
cd ..
./gradlew :listings-service:run
```
Run each of the other three services the same way, in separate terminals (or background them with `&`). Default ports: listings `8081`, search `8082`, messaging `8083`, payments `8084`.

## Option B: full local k8s (kind)

Best for seeing the actual container-diagram topology running, and for experimenting with multiple replicas (see `infra/k8s/12-messaging.yaml`). Requires **Helm 3** in addition to kind/kubectl (Gloo + Flagger install).

```bash
./scripts/build-images.sh     # gradle installDist + docker build, all 4 services
./scripts/deploy-kind.sh      # kind create cluster, load images, LGTM, Gloo+Flagger, apps
kubectl -n c2c get pods -w    # watch until everything is Running/Ready
kubectl -n c2c get canary     # listings-service should be Initialized/Succeeded
```

- **Listings** (`localhost:8081`) go through **Gloo Edge** gateway-proxy (NodePort 30081). Prefer `curl -H 'Host: listings.local' http://localhost:8081/...` (domain `*` is also configured so bare curls usually work).
- **Search / messaging / payments** stay on direct NodePorts `8082`–`8084`.
- **Grafana:** http://localhost:3000 (anonymous Viewer).

Istio fallback (if Gloo is painful locally):

```bash
PROGRESSIVE_PROVIDER=istio ./scripts/deploy-kind.sh
```

To tear down: `kind delete cluster --name c2c-marketplace`.

### Progressive delivery (listings canary)

Flagger analyzes and promotes `listings-service` using Prometheus at `http://prometheus.c2c:9090` (existing LGTM stack — Datadog stand-in). Design notes: [`docs/plans/2026-07-20-flagger-gloo-progressive-delivery-design.md`](docs/plans/2026-07-20-flagger-gloo-progressive-delivery-design.md).

```bash
./scripts/canary-listings.sh
# or: IMAGE_TAG=v2 ./scripts/canary-listings.sh
# watch-only: ./scripts/canary-listings.sh --watch-only
```

During `Progressing`, Flagger steps canary weight (10% → 50%), checks Micrometer success-rate ≥ 99% and p99 ≤ 500ms, then promotes to primary. To demo rollback, generate 5xx against the canary while analysis is running (or crash the canary pods) until the failure threshold is hit.

## Exercising the end-to-end flow

**1. Create a listing** (this also runs it through the trust & safety stub and publishes `listing.created`):
```bash
curl -X POST localhost:8081/listings -H 'Content-Type: application/json' -d '{
  "sellerId": "seller-1",
  "title": "Mid-century desk",
  "description": "Solid wood, minor scuffs",
  "priceCents": 4000,
  "category": "furniture",
  "lat": 47.6062,
  "lon": -122.3321
}'
```

**2. Search for it** (search-service consumes the Kafka event async — if this returns nothing immediately, wait a second and retry; that lag *is* the eventual-consistency trade-off described in the TDD):
```bash
curl "localhost:8082/search?q=desk&lat=47.6&lon=-122.3&radiusKm=25"
```

**3. Chat about it** (WebSocket — use `websocat` or `wscat`; conversation IDs in this mock are `<userA>:<userB>`):
```bash
# terminal 1
websocat "ws://localhost:8083/ws/buyer-1"
# terminal 2
websocat "ws://localhost:8083/ws/seller-1"
# in terminal 1, send:
{"conversationId": "buyer-1:seller-1", "body": "Is this still available?"}
# it should arrive in terminal 2 almost immediately
```

**4. Buy it** (creates the order + escrow hold in one transaction):
```bash
curl -X POST localhost:8084/orders -H 'Content-Type: application/json' -d '{
  "listingId": "<listing id from step 1>",
  "buyerId": "buyer-1",
  "sellerId": "seller-1",
  "amountCents": 4000
}'
```

**5. Confirm delivery** (releases escrow — the transition is validated by `EscrowStateMachine`, see its unit tests):
```bash
curl -X POST "localhost:8084/orders/<order id from step 4>/confirm-delivery"
```

## Synthetic data

Reproducible demo / light-load traffic against localhost services (compose or kind). Requires the stack up, `jq`, and `websocat` (for chat). Design notes: [`docs/plans/2026-07-12-synth-data-factory-design.md`](docs/plans/2026-07-12-synth-data-factory-design.md). Cursor orchestration agent: [`.claude/agents/synth-data-factory.md`](.claude/agents/synth-data-factory.md).

```bash
./scripts/synth-run.sh demo         # ~10 listings, mix confirm/dispute, one chat pair
./scripts/synth-run.sh load-light   # ~100 listings, mild load
```

## Running tests

```bash
./gradlew :payments-service:test    # pure escrow state machine — no Docker needed
./gradlew :listings-service:test    # Testcontainers-based, needs Docker running
./gradlew :synth-harness:test       # synthetic data generators / harness unit tests
```

## What to poke at next

- Scale `messaging-service` to 2+ replicas (already set in the k8s manifest) and watch `kubectl logs` on both pods while chatting — you should see the Redis pub/sub hand-off in `ConnectionRegistry` fire when sender and recipient land on different pods.
- Kill the `search-service` pod mid-index and watch it resume from its committed Kafka offset instead of re-indexing everything or losing events.
- Try to make `EscrowStateMachine` do something illegal (e.g. dispute an already-released order) and confirm it throws rather than silently corrupting state — then try the same thing against the HTTP API and check you get a `409`, not a `500`.

See `docs/TDD.md` section 9 for what's deliberately left unbuilt (auth, full multi-service gateway, etc.) and why. Listings canary + LGTM observability are implemented for kind demos.
