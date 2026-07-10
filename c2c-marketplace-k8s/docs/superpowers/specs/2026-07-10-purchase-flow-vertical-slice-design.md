# Purchase Flow Vertical Slice — TDD Design

**Date:** 2026-07-10
**Author:** Terence Schumacher
**Status:** Approved

## 1. Scope

End-to-end vertical slice of the C2C marketplace purchase flow, implemented and verified via TDD (inside-out), and demonstrated live on both docker-compose and a local kind cluster.

**Slice:** `POST /listings` → `POST /orders` → `POST /orders/{id}/confirm-delivery` → escrow status = RELEASED

This is the core buyer-protection flow: a seller creates a listing, a buyer places an order (payment held in escrow), the buyer confirms delivery, and the escrow is released to the seller.

**Out of scope for this slice:**
- search-service (listing-to-search flow is a separate slice)
- messaging-service (WebSocket flow is a separate slice)
- Kafka event publishing verification (EventPublisher is best-effort async; not tested in this slice)
- Dispute/refund path (covered by `EscrowStateMachineTest`; not included in smoke test)

---

## 2. What Already Exists (Unchanged)

| File | Status |
|---|---|
| `EscrowStateMachineTest` | Complete — exhaustive over all 9 (state, event) pairs |
| `ListingRepositoryTest` | Complete — create/find/sold with real Postgres via Testcontainers |
| `listings-service` production code | Complete |
| `payments-service` production code | Complete |
| `infra/docker-compose.yml` | Complete |
| `infra/k8s/` manifests | Complete |
| `scripts/build-images.sh`, `scripts/deploy-kind.sh` | Complete |

---

## 3. Test Architecture

Four layers, red → green → refactor at each layer before moving to the next.

| Layer | Class/File | Mechanism | Infra Required |
|---|---|---|---|
| 1 | `OrderRepositoryTest` | Testcontainers (real Postgres) | Docker |
| 2 | `PaymentsRoutesTest` | Ktor test engine, mock repository | None |
| 3 | `ListingRoutesTest` | Ktor test engine, mock repository + publisher | None |
| 4 | `scripts/smoke-test.sh` | curl against live running stack | docker-compose or kind |

---

## 4. Test Cases

### Layer 1 — `OrderRepositoryTest`

Located at: `payments-service/src/test/kotlin/com/marketplace/payments/OrderRepositoryTest.kt`

| Test | Assertion |
|---|---|
| `createWithHold persists order and escrow_hold atomically` | Both rows committed; order status = HELD, escrow status = HELD |
| `applyEvent HELD + ConfirmDelivery → RELEASED` | Escrow row updated to RELEASED; method returns RELEASED |
| `applyEvent HELD + BuyerDispute → REFUNDED` | Escrow row updated to REFUNDED; method returns REFUNDED |
| `applyEvent on non-existent order throws NoSuchElementException` | Exception thrown; no DB mutation |
| `applyEvent on already-RELEASED order throws IllegalEscrowTransitionException` | State machine violation propagated; no DB mutation |

Setup: `PostgreSQLContainer("postgres:16-alpine")` started in `@BeforeAll`, schema created via Exposed `SchemaUtils.createMissingTablesAndColumns`.

### Layer 2 — `PaymentsRoutesTest`

Located at: `payments-service/src/test/kotlin/com/marketplace/payments/PaymentsRoutesTest.kt`

Uses Ktor `testApplication {}` with a mock `OrderRepository`. `EventPublisher` is also mocked (no-op).

| Test | HTTP assertion |
|---|---|
| `POST /orders → 201` | Body contains `orderId` and `status = HELD` |
| `POST /orders/{id}/confirm-delivery → 200` | Body contains `status = RELEASED` |
| `POST /orders/{id}/dispute → 200` | Body contains `status = REFUNDED` |
| `POST /orders/{released-id}/confirm-delivery → 409` | `IllegalEscrowTransitionException` maps to Conflict |
| `POST /orders/{unknown}/confirm-delivery → 404` | `NoSuchElementException` maps to Not Found |

### Layer 3 — `ListingRoutesTest`

Located at: `listings-service/src/test/kotlin/com/marketplace/listings/ListingRoutesTest.kt`

Uses Ktor `testApplication {}` with a mock `ListingRepository` and mock `EventPublisher`.

| Test | HTTP assertion |
|---|---|
| `POST /listings → 201` | Body contains listing fields; publisher called once |
| `POST /listings with blocked keyword → 422` | T&S gate fires; repository and publisher not called |
| `GET /listings/{id} → 200` | Returns listing body |
| `GET /listings/{unknown} → 404` | Returns error body |

### Layer 4 — `scripts/smoke-test.sh`

Sequential curl commands with `jq` assertions. Accepts env vars `LISTINGS_URL` (default `http://localhost:8081`) and `PAYMENTS_URL` (default `http://localhost:8084`).

Steps:
1. `POST $LISTINGS_URL/listings` — assert HTTP 201, capture `listingId` from JSON
2. `POST $PAYMENTS_URL/orders` (buyerId, sellerId, listingId, amountCents) — assert HTTP 201, capture `orderId`, assert `status == "HELD"`
3. `POST $PAYMENTS_URL/orders/$orderId/confirm-delivery` — assert HTTP 200, assert `status == "RELEASED"`
4. Print `PASS` with elapsed time, or `FAIL` with the step that failed and the actual response

The same script runs against docker-compose (default URLs) and kind (`kubectl port-forward` to the same ports before running).

---

## 5. Build Configuration Change

`payments-service/build.gradle.kts` needs Testcontainers and Ktor test engine dependencies added, mirroring `listings-service/build.gradle.kts`.

Dependencies to add:
- `testImplementation("org.testcontainers:postgresql")`
- `testImplementation("org.testcontainers:junit-jupiter")`
- `testImplementation("io.ktor:ktor-server-test-host")`
- `testImplementation("io.ktor:ktor-client-content-negotiation")`

`listings-service/build.gradle.kts` also needs Ktor test engine dependencies for `ListingRoutesTest` (Testcontainers is already present).

---

## 6. Production Code Status

No production code gaps. `OrderRepository.applyEvent` already throws `NoSuchElementException` when the order is not found (`OrderRepository.kt:101`), which the route correctly maps to 404. `EscrowStateMachine.transition` already throws `IllegalEscrowTransitionException` for illegal transitions, which the route maps to 409. The TDD work in this slice is entirely in the test layer.

---

## 7. Deliverable State

When done, the following commands complete successfully in sequence:

```bash
# All tests green
./gradlew test

# docker-compose live demo
cd infra && docker compose up -d
cd ..
./gradlew :listings-service:run &
./gradlew :payments-service:run &
sleep 5
./scripts/smoke-test.sh   # prints PASS

# kind cluster live demo
./scripts/build-images.sh
./scripts/deploy-kind.sh
kubectl -n c2c port-forward svc/listings 8081:8081 &
kubectl -n c2c port-forward svc/payments 8084:8084 &
sleep 10
./scripts/smoke-test.sh   # prints PASS

kind delete cluster --name c2c-marketplace
```

---

## 8. TDD Cadence

For each layer:
1. Write one failing test (red)
2. Write the minimum code to make it pass (green)
3. Refactor if needed
4. Repeat for the next test case in that layer
5. When all cases in a layer are green, move to the next layer

The smoke test (Layer 4) is written after all Gradle tests are green. It is the capstone, not a driver.
