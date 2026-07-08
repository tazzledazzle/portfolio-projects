# C2C marketplace (mock/simulated)

A runnable, simplified simulation of a mid-scale mobile-first C2C marketplace — four Kotlin/Ktor services (listings, search, messaging, payments) backed by Postgres, Redis, OpenSearch, and Kafka (via Redpanda), deployable to a local kind cluster.

Full design rationale, data flow diagrams, and known simplifications: [`docs/TDD.md`](docs/TDD.md).

## Prerequisites

- JDK 17+
- Docker (with enough memory allocated — OpenSearch alone wants ~1GB)
- [kind](https://kind.sigs.k8s.io/) and `kubectl`, for the full k8s path
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

Best for seeing the actual container-diagram topology running, and for experimenting with multiple replicas (see `infra/k8s/12-messaging.yaml`).

```bash
./scripts/build-images.sh     # gradle installDist + docker build, all 4 services
./scripts/deploy-kind.sh      # kind create cluster, load images, kubectl apply
kubectl -n c2c get pods -w    # watch until everything is Running/Ready
```

Services are reachable at `localhost:8081`–`8084` via kind's `extraPortMappings` (no `kubectl port-forward` needed for basic testing).

To tear down: `kind delete cluster --name c2c-marketplace`.

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

## Running tests

```bash
./gradlew :payments-service:test    # pure escrow state machine — no Docker needed
./gradlew :listings-service:test    # Testcontainers-based, needs Docker running
```

## What to poke at next

- Scale `messaging-service` to 2+ replicas (already set in the k8s manifest) and watch `kubectl logs` on both pods while chatting — you should see the Redis pub/sub hand-off in `ConnectionRegistry` fire when sender and recipient land on different pods.
- Kill the `search-service` pod mid-index and watch it resume from its committed Kafka offset instead of re-indexing everything or losing events.
- Try to make `EscrowStateMachine` do something illegal (e.g. dispute an already-released order) and confirm it throws rather than silently corrupting state — then try the same thing against the HTTP API and check you get a `409`, not a `500`.

See `docs/TDD.md` section 9 for what's deliberately left unbuilt (auth, API gateway, observability) and why.
