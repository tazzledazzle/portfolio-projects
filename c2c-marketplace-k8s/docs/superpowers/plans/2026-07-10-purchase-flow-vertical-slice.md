# Purchase Flow Vertical Slice Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a TDD-verified test suite for the purchase flow vertical slice (create listing → place order → confirm delivery → escrow released) and prove it live on both docker-compose and kind.

**Architecture:** Inside-out TDD — `OrderRepositoryTest` (Testcontainers real Postgres) first, then `PaymentsRoutesTest` (Ktor test engine + Mockk), then `ListingRoutesTest` (Ktor test engine + Mockk), then a `smoke-test.sh` script that drives the live stack curl-by-curl. All production code is already implemented; this plan adds tests only (plus `smoke-test.sh`).

**Tech Stack:** Kotlin/JVM 17, Ktor 2.3.12, Exposed 0.50.1, Testcontainers 1.19.8, JUnit 5, MockK 1.13.11, Postgres 16-alpine, Gradle 8.7, kind, kubectl, curl, jq.

## Global Constraints

- Kotlin JVM target: 17
- Ktor version: 2.3.12 (match existing `ktorVersion` variable in build files)
- Testcontainers version: 1.19.8 (match existing version in listings-service)
- MockK version: 1.13.11
- Do NOT modify any production source files (`src/main/`)
- Do NOT modify `infra/docker-compose.yml` or any `infra/k8s/` manifests
- `smoke-test.sh` must be executable and accept `LISTINGS_URL` / `PAYMENTS_URL` env vars

---

## File Map

| File | Action | Responsibility |
|---|---|---|
| `payments-service/build.gradle.kts` | Modify | Add Testcontainers, Ktor test engine, MockK test deps |
| `payments-service/src/test/kotlin/com/marketplace/payments/OrderRepositoryTest.kt` | Create | Testcontainers tests for the two-write atomic transaction |
| `payments-service/src/test/kotlin/com/marketplace/payments/PaymentsRoutesTest.kt` | Create | Ktor test engine tests for the payments HTTP surface |
| `listings-service/build.gradle.kts` | Modify | Add ktor-client-content-negotiation and MockK test deps |
| `listings-service/src/test/kotlin/com/marketplace/listings/ListingRoutesTest.kt` | Create | Ktor test engine tests for the listings HTTP surface |
| `scripts/smoke-test.sh` | Create | E2E curl script for docker-compose and kind verification |

---

## Task 1: OrderRepositoryTest

**Files:**
- Modify: `payments-service/build.gradle.kts`
- Create: `payments-service/src/test/kotlin/com/marketplace/payments/OrderRepositoryTest.kt`

**Interfaces:**
- Consumes: `OrderRepository`, `OrderTable`, `EscrowHoldTable`, `CreateOrderRequest`, `Order`, `EscrowStatus`, `EscrowEvent`, `IllegalEscrowTransitionException` from `payments-service/src/main/`
- Produces: nothing (test-only)

- [ ] **Step 1: Add test dependencies to payments-service/build.gradle.kts**

Replace the `dependencies` block's test section. The full updated file:

```kotlin
plugins {
    kotlin("jvm")
    kotlin("plugin.serialization")
    application
}

application {
    mainClass.set("com.marketplace.payments.ApplicationKt")
}

dependencies {
    implementation(project(":common"))

    val ktorVersion = "2.3.12"
    implementation("io.ktor:ktor-server-core:$ktorVersion")
    implementation("io.ktor:ktor-server-netty:$ktorVersion")
    implementation("io.ktor:ktor-server-content-negotiation:$ktorVersion")
    implementation("io.ktor:ktor-serialization-kotlinx-json:$ktorVersion")
    implementation("io.ktor:ktor-server-status-pages:$ktorVersion")
    implementation("io.ktor:ktor-server-call-logging:$ktorVersion")

    implementation("org.jetbrains.exposed:exposed-core:0.50.1")
    implementation("org.jetbrains.exposed:exposed-dao:0.50.1")
    implementation("org.jetbrains.exposed:exposed-jdbc:0.50.1")
    implementation("org.jetbrains.exposed:exposed-java-time:0.50.1")
    implementation("org.postgresql:postgresql:42.7.3")
    implementation("com.zaxxer:HikariCP:5.1.0")

    implementation("org.apache.kafka:kafka-clients:3.7.0")
    implementation("org.jetbrains.kotlinx:kotlinx-serialization-json:1.6.3")

    implementation("ch.qos.logback:logback-classic:1.5.6")

    testImplementation(kotlin("test"))
    testImplementation("org.junit.jupiter:junit-jupiter-params:5.10.2")
    testImplementation("io.ktor:ktor-server-test-host:$ktorVersion")
    testImplementation("io.ktor:ktor-client-content-negotiation:$ktorVersion")
    testImplementation("org.testcontainers:postgresql:1.19.8")
    testImplementation("org.testcontainers:junit-jupiter:1.19.8")
    testImplementation("io.mockk:mockk:1.13.11")
}

kotlin {
    jvmToolchain(17)
}

tasks.test {
    useJUnitPlatform()
}
```

- [ ] **Step 2: Verify the build resolves**

```bash
./gradlew :payments-service:build -x test
```

Expected: `BUILD SUCCESSFUL`

- [ ] **Step 3: Write the first failing test — createWithHold**

Create `payments-service/src/test/kotlin/com/marketplace/payments/OrderRepositoryTest.kt`:

```kotlin
package com.marketplace.payments

import org.jetbrains.exposed.sql.Database
import org.jetbrains.exposed.sql.SchemaUtils
import org.jetbrains.exposed.sql.transactions.transaction
import org.junit.jupiter.api.AfterAll
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Assertions.assertNotNull
import org.junit.jupiter.api.Assertions.assertThrows
import org.junit.jupiter.api.BeforeAll
import org.junit.jupiter.api.Test
import org.testcontainers.containers.PostgreSQLContainer
import org.testcontainers.junit.jupiter.Testcontainers

@Testcontainers
class OrderRepositoryTest {

    companion object {
        private val postgres = PostgreSQLContainer("postgres:16-alpine")
            .withDatabaseName("marketplace_test")
            .withUsername("test")
            .withPassword("test")

        @JvmStatic
        @BeforeAll
        fun setup() {
            postgres.start()
            Database.connect(
                url = postgres.jdbcUrl,
                user = postgres.username,
                password = postgres.password
            )
            transaction { SchemaUtils.createMissingTablesAndColumns(OrderTable, EscrowHoldTable) }
        }

        @JvmStatic
        @AfterAll
        fun teardown() {
            postgres.stop()
        }
    }

    private val repository = OrderRepository()

    private fun sampleRequest() = CreateOrderRequest(
        listingId = "listing-1",
        buyerId = "buyer-1",
        sellerId = "seller-1",
        amountCents = 5000
    )

    @Test
    fun `createWithHold persists order and escrow hold atomically`() {
        val order = repository.createWithHold(sampleRequest())

        assertNotNull(order.id)
        assertEquals("HELD", order.status)
        assertEquals(EscrowStatus.HELD, repository.currentEscrowStatus(order.id))
    }
}
```

- [ ] **Step 4: Run the test — expect green (production code already exists)**

```bash
./gradlew :payments-service:test --tests "com.marketplace.payments.OrderRepositoryTest.createWithHold*" --info
```

Expected: `1 test completed` — if it fails, the most likely cause is a missing JDBC driver on the classpath; check that `postgresql` is in `implementation` deps.

- [ ] **Step 5: Add applyEvent transition tests**

Append to `OrderRepositoryTest.kt` (inside the class, after the first test):

```kotlin
    @Test
    fun `applyEvent ConfirmDelivery transitions HELD to RELEASED`() {
        val order = repository.createWithHold(sampleRequest())

        val result = repository.applyEvent(order.id, EscrowEvent.ConfirmDelivery)

        assertEquals(EscrowStatus.RELEASED, result)
        assertEquals(EscrowStatus.RELEASED, repository.currentEscrowStatus(order.id))
    }

    @Test
    fun `applyEvent BuyerDispute transitions HELD to REFUNDED`() {
        val order = repository.createWithHold(sampleRequest())

        val result = repository.applyEvent(order.id, EscrowEvent.BuyerDispute)

        assertEquals(EscrowStatus.REFUNDED, result)
        assertEquals(EscrowStatus.REFUNDED, repository.currentEscrowStatus(order.id))
    }
```

- [ ] **Step 6: Run transition tests**

```bash
./gradlew :payments-service:test --tests "com.marketplace.payments.OrderRepositoryTest" --info
```

Expected: `3 tests completed`

- [ ] **Step 7: Add error case tests**

Append to `OrderRepositoryTest.kt` (inside the class):

```kotlin
    @Test
    fun `applyEvent on non-existent order throws NoSuchElementException`() {
        assertThrows(NoSuchElementException::class.java) {
            repository.applyEvent("does-not-exist", EscrowEvent.ConfirmDelivery)
        }
    }

    @Test
    fun `applyEvent on already RELEASED order throws IllegalEscrowTransitionException`() {
        val order = repository.createWithHold(sampleRequest())
        repository.applyEvent(order.id, EscrowEvent.ConfirmDelivery) // HELD → RELEASED

        assertThrows(IllegalEscrowTransitionException::class.java) {
            repository.applyEvent(order.id, EscrowEvent.ConfirmDelivery) // RELEASED → illegal
        }
    }
```

- [ ] **Step 8: Run all five tests**

```bash
./gradlew :payments-service:test --tests "com.marketplace.payments.OrderRepositoryTest" --info
```

Expected: `5 tests completed, 5 passed`

- [ ] **Step 9: Commit**

```bash
git add payments-service/build.gradle.kts \
        payments-service/src/test/kotlin/com/marketplace/payments/OrderRepositoryTest.kt
git commit -m "test: add OrderRepositoryTest — Testcontainers, 5 cases"
```

---

## Task 2: PaymentsRoutesTest

**Files:**
- Create: `payments-service/src/test/kotlin/com/marketplace/payments/PaymentsRoutesTest.kt`

**Interfaces:**
- Consumes: `Application.module(OrderRepository, EventPublisher)` from `Application.kt`; `Order`, `CreateOrderRequest`, `ErrorResponse`, `EscrowEvent`, `EscrowStatus`, `IllegalEscrowTransitionException` from production source
- Produces: nothing (test-only)

**Note on MockK and final classes:** Kotlin classes are `final` by default. MockK 1.13.x uses ByteBuddy instrumentation to mock final classes without requiring the `open` keyword. `mockk<OrderRepository>()` and `mockk<EventPublisher>(relaxed = true)` work as-is. `relaxed = true` means Unit-returning methods (`publishOrderCreated`, `publishOrderCompleted`) are no-ops unless explicitly stubbed.

- [ ] **Step 1: Write the happy-path tests**

Create `payments-service/src/test/kotlin/com/marketplace/payments/PaymentsRoutesTest.kt`:

```kotlin
package com.marketplace.payments

import io.ktor.client.call.body
import io.ktor.client.plugins.contentnegotiation.ContentNegotiation
import io.ktor.client.request.post
import io.ktor.client.request.setBody
import io.ktor.http.ContentType
import io.ktor.http.HttpStatusCode
import io.ktor.http.contentType
import io.ktor.serialization.kotlinx.json.json
import io.ktor.server.testing.testApplication
import io.mockk.every
import io.mockk.mockk
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Test

class PaymentsRoutesTest {

    private fun makeOrder(id: String = "order-1", status: String = "HELD") =
        Order(
            id = id,
            listingId = "listing-1",
            buyerId = "buyer-1",
            sellerId = "seller-1",
            amountCents = 5000,
            status = status
        )

    @Test
    fun `POST orders returns 201 with orderId and status HELD`() = testApplication {
        val repository = mockk<OrderRepository>()
        val publisher = mockk<EventPublisher>(relaxed = true)
        every { repository.createWithHold(any()) } returns makeOrder()

        application { module(repository, publisher) }
        val client = createClient { install(ContentNegotiation) { json() } }

        val response = client.post("/orders") {
            contentType(ContentType.Application.Json)
            setBody(CreateOrderRequest("listing-1", "buyer-1", "seller-1", 5000))
        }

        assertEquals(HttpStatusCode.Created, response.status)
        val body = response.body<Order>()
        assertEquals("order-1", body.id)
        assertEquals("HELD", body.status)
    }

    @Test
    fun `POST orders id confirm-delivery returns 200 with status RELEASED`() = testApplication {
        val repository = mockk<OrderRepository>()
        val publisher = mockk<EventPublisher>(relaxed = true)
        every { repository.applyEvent("order-1", EscrowEvent.ConfirmDelivery) } returns EscrowStatus.RELEASED

        application { module(repository, publisher) }
        val client = createClient { install(ContentNegotiation) { json() } }

        val response = client.post("/orders/order-1/confirm-delivery")

        assertEquals(HttpStatusCode.OK, response.status)
        val body = response.body<Map<String, String>>()
        assertEquals("RELEASED", body["status"])
    }

    @Test
    fun `POST orders id dispute returns 200 with status REFUNDED`() = testApplication {
        val repository = mockk<OrderRepository>()
        val publisher = mockk<EventPublisher>(relaxed = true)
        every { repository.applyEvent("order-1", EscrowEvent.BuyerDispute) } returns EscrowStatus.REFUNDED

        application { module(repository, publisher) }
        val client = createClient { install(ContentNegotiation) { json() } }

        val response = client.post("/orders/order-1/dispute")

        assertEquals(HttpStatusCode.OK, response.status)
        val body = response.body<Map<String, String>>()
        assertEquals("REFUNDED", body["status"])
    }
}
```

- [ ] **Step 2: Run the three happy-path tests**

```bash
./gradlew :payments-service:test --tests "com.marketplace.payments.PaymentsRoutesTest" --info
```

Expected: `3 tests completed, 3 passed`. If you see `ClassNotFoundException` for `io.mockk.*`, confirm the MockK dependency was added in Task 1.

- [ ] **Step 3: Add the error case tests**

Append to `PaymentsRoutesTest.kt` (inside the class):

```kotlin
    @Test
    fun `POST orders id confirm-delivery returns 409 when order already released`() = testApplication {
        val repository = mockk<OrderRepository>()
        val publisher = mockk<EventPublisher>(relaxed = true)
        every { repository.applyEvent(any(), EscrowEvent.ConfirmDelivery) } throws
            IllegalEscrowTransitionException("cannot apply ConfirmDelivery to order already in state RELEASED")

        application { module(repository, publisher) }
        val client = createClient { install(ContentNegotiation) { json() } }

        val response = client.post("/orders/order-1/confirm-delivery")

        assertEquals(HttpStatusCode.Conflict, response.status)
    }

    @Test
    fun `POST orders id confirm-delivery returns 404 when order not found`() = testApplication {
        val repository = mockk<OrderRepository>()
        val publisher = mockk<EventPublisher>(relaxed = true)
        every { repository.applyEvent(any(), EscrowEvent.ConfirmDelivery) } throws
            NoSuchElementException("no escrow hold for order unknown")

        application { module(repository, publisher) }
        val client = createClient { install(ContentNegotiation) { json() } }

        val response = client.post("/orders/unknown/confirm-delivery")

        assertEquals(HttpStatusCode.NotFound, response.status)
    }
```

- [ ] **Step 4: Run all five payments route tests**

```bash
./gradlew :payments-service:test --tests "com.marketplace.payments.PaymentsRoutesTest" --info
```

Expected: `5 tests completed, 5 passed`

- [ ] **Step 5: Run the full payments-service test suite**

```bash
./gradlew :payments-service:test --info
```

Expected: `8 tests completed` (3 from `EscrowStateMachineTest` counting the parameterized cases, plus 2 from the non-parameterized + 5 from `OrderRepositoryTest` + 5 from `PaymentsRoutesTest` = at least 14 tests). Confirm `0 failures`.

- [ ] **Step 6: Commit**

```bash
git add payments-service/src/test/kotlin/com/marketplace/payments/PaymentsRoutesTest.kt
git commit -m "test: add PaymentsRoutesTest — Ktor engine, 5 HTTP surface cases"
```

---

## Task 3: ListingRoutesTest

**Files:**
- Modify: `listings-service/build.gradle.kts`
- Create: `listings-service/src/test/kotlin/com/marketplace/listings/ListingRoutesTest.kt`

**Interfaces:**
- Consumes: `Application.module(ListingRepository, EventPublisher)` from `listings-service/src/main/kotlin/.../Application.kt`; `Listing`, `CreateListingRequest`, `ErrorResponse` from production source
- Produces: nothing (test-only)

- [ ] **Step 1: Add test dependencies to listings-service/build.gradle.kts**

The `ktor-server-test-host` and Testcontainers are already present. Add `ktor-client-content-negotiation` and MockK. Replace the test dependency lines at the bottom of the `dependencies` block:

```kotlin
    testImplementation(kotlin("test"))
    testImplementation("io.ktor:ktor-server-test-host:$ktorVersion")
    testImplementation("io.ktor:ktor-client-content-negotiation:$ktorVersion")
    testImplementation("org.testcontainers:postgresql:1.19.8")
    testImplementation("org.testcontainers:junit-jupiter:1.19.8")
    testImplementation("io.mockk:mockk:1.13.11")
```

Full updated `listings-service/build.gradle.kts`:

```kotlin
plugins {
    kotlin("jvm")
    kotlin("plugin.serialization")
    application
}

application {
    mainClass.set("com.marketplace.listings.ApplicationKt")
}

dependencies {
    implementation(project(":common"))

    val ktorVersion = "2.3.12"
    implementation("io.ktor:ktor-server-core:$ktorVersion")
    implementation("io.ktor:ktor-server-netty:$ktorVersion")
    implementation("io.ktor:ktor-server-content-negotiation:$ktorVersion")
    implementation("io.ktor:ktor-serialization-kotlinx-json:$ktorVersion")
    implementation("io.ktor:ktor-server-status-pages:$ktorVersion")
    implementation("io.ktor:ktor-server-call-logging:$ktorVersion")

    implementation("org.jetbrains.exposed:exposed-core:0.50.1")
    implementation("org.jetbrains.exposed:exposed-dao:0.50.1")
    implementation("org.jetbrains.exposed:exposed-jdbc:0.50.1")
    implementation("org.jetbrains.exposed:exposed-java-time:0.50.1")
    implementation("org.postgresql:postgresql:42.7.3")
    implementation("com.zaxxer:HikariCP:5.1.0")

    implementation("org.apache.kafka:kafka-clients:3.7.0")
    implementation("org.jetbrains.kotlinx:kotlinx-serialization-json:1.6.3")

    implementation("ch.qos.logback:logback-classic:1.5.6")

    testImplementation(kotlin("test"))
    testImplementation("io.ktor:ktor-server-test-host:$ktorVersion")
    testImplementation("io.ktor:ktor-client-content-negotiation:$ktorVersion")
    testImplementation("org.testcontainers:postgresql:1.19.8")
    testImplementation("org.testcontainers:junit-jupiter:1.19.8")
    testImplementation("io.mockk:mockk:1.13.11")
}

kotlin {
    jvmToolchain(17)
}

tasks.test {
    useJUnitPlatform()
}
```

- [ ] **Step 2: Verify build resolves**

```bash
./gradlew :listings-service:build -x test
```

Expected: `BUILD SUCCESSFUL`

- [ ] **Step 3: Write the happy-path and T&S tests**

Create `listings-service/src/test/kotlin/com/marketplace/listings/ListingRoutesTest.kt`:

```kotlin
package com.marketplace.listings

import io.ktor.client.call.body
import io.ktor.client.plugins.contentnegotiation.ContentNegotiation
import io.ktor.client.request.get
import io.ktor.client.request.post
import io.ktor.client.request.setBody
import io.ktor.http.ContentType
import io.ktor.http.HttpStatusCode
import io.ktor.http.contentType
import io.ktor.serialization.kotlinx.json.json
import io.ktor.server.testing.testApplication
import io.mockk.every
import io.mockk.mockk
import io.mockk.verify
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Test

class ListingRoutesTest {

    private fun fakeListing(id: String = "listing-1") = Listing(
        id = id,
        sellerId = "seller-1",
        title = "Test Bike",
        description = null,
        priceCents = 5000,
        category = "sporting-goods",
        lat = 47.6062,
        lon = -122.3321,
        status = "ACTIVE",
        createdAtEpochMillis = 1_700_000_000_000L
    )

    @Test
    fun `POST listings returns 201 with listing body and publishes event`() = testApplication {
        val repository = mockk<ListingRepository>()
        val publisher = mockk<EventPublisher>(relaxed = true)
        every { repository.create(any()) } returns fakeListing()

        application { module(repository, publisher) }
        val client = createClient { install(ContentNegotiation) { json() } }

        val response = client.post("/listings") {
            contentType(ContentType.Application.Json)
            setBody(
                CreateListingRequest(
                    sellerId = "seller-1",
                    title = "Test Bike",
                    priceCents = 5000,
                    category = "sporting-goods",
                    lat = 47.6062,
                    lon = -122.3321
                )
            )
        }

        assertEquals(HttpStatusCode.Created, response.status)
        val body = response.body<Listing>()
        assertEquals("listing-1", body.id)
        assertEquals("ACTIVE", body.status)
        verify(exactly = 1) { publisher.publishListingCreated(any()) }
    }

    @Test
    fun `POST listings with blocked keyword returns 422 without touching repository`() = testApplication {
        val repository = mockk<ListingRepository>(relaxed = true)
        val publisher = mockk<EventPublisher>(relaxed = true)

        application { module(repository, publisher) }
        val client = createClient { install(ContentNegotiation) { json() } }

        val response = client.post("/listings") {
            contentType(ContentType.Application.Json)
            setBody(
                CreateListingRequest(
                    sellerId = "seller-1",
                    title = "Stolen bike for sale",
                    priceCents = 5000,
                    category = "sporting-goods",
                    lat = 47.6062,
                    lon = -122.3321
                )
            )
        }

        assertEquals(HttpStatusCode.UnprocessableEntity, response.status)
        verify(exactly = 0) { repository.create(any()) }
        verify(exactly = 0) { publisher.publishListingCreated(any()) }
    }

    @Test
    fun `GET listings id returns 200 with listing body`() = testApplication {
        val repository = mockk<ListingRepository>()
        val publisher = mockk<EventPublisher>(relaxed = true)
        every { repository.findById("listing-1") } returns fakeListing()

        application { module(repository, publisher) }
        val client = createClient { install(ContentNegotiation) { json() } }

        val response = client.get("/listings/listing-1")

        assertEquals(HttpStatusCode.OK, response.status)
        val body = response.body<Listing>()
        assertEquals("listing-1", body.id)
    }

    @Test
    fun `GET listings id returns 404 when listing not found`() = testApplication {
        val repository = mockk<ListingRepository>()
        val publisher = mockk<EventPublisher>(relaxed = true)
        every { repository.findById("unknown") } returns null

        application { module(repository, publisher) }
        val client = createClient { install(ContentNegotiation) { json() } }

        val response = client.get("/listings/unknown")

        assertEquals(HttpStatusCode.NotFound, response.status)
    }
}
```

- [ ] **Step 4: Run all four listing route tests**

```bash
./gradlew :listings-service:test --tests "com.marketplace.listings.ListingRoutesTest" --info
```

Expected: `4 tests completed, 4 passed`

- [ ] **Step 5: Run the full listings-service test suite**

```bash
./gradlew :listings-service:test --info
```

Expected: All tests pass (3 from `ListingRepositoryTest` + 4 from `ListingRoutesTest` = 7 total). Confirm `0 failures`.

- [ ] **Step 6: Run the complete Gradle test suite**

```bash
./gradlew test
```

Expected: All modules pass. Confirm `BUILD SUCCESSFUL` with `0 failures`.

- [ ] **Step 7: Commit**

```bash
git add listings-service/build.gradle.kts \
        listings-service/src/test/kotlin/com/marketplace/listings/ListingRoutesTest.kt
git commit -m "test: add ListingRoutesTest — Ktor engine, 4 HTTP surface cases"
```

---

## Task 4: smoke-test.sh + docker-compose live demo

**Files:**
- Create: `scripts/smoke-test.sh`

**Interfaces:**
- Consumes: `POST /listings` on `LISTINGS_URL`, `POST /orders` and `POST /orders/{id}/confirm-delivery` on `PAYMENTS_URL`
- Produces: exit code 0 (PASS) or 1 (FAIL with step name and actual response)

**Prerequisites:** `jq` must be installed (`brew install jq` on macOS).

- [ ] **Step 1: Create scripts/smoke-test.sh**

```bash
#!/usr/bin/env bash
set -euo pipefail

LISTINGS_URL="${LISTINGS_URL:-http://localhost:8081}"
PAYMENTS_URL="${PAYMENTS_URL:-http://localhost:8084}"
START=$(date +%s)

wait_for() {
    local url="$1" name="$2" n=0 max=30
    printf "Waiting for %s" "$name"
    while ! curl -sf "$url/healthz" > /dev/null 2>&1; do
        n=$((n + 1))
        [ $n -ge $max ] && { echo " TIMEOUT"; echo "FAIL: $name did not become healthy after ${max}s"; exit 1; }
        printf "."
        sleep 1
    done
    echo " ready"
}

wait_for "$LISTINGS_URL" "listings-service"
wait_for "$PAYMENTS_URL" "payments-service"

echo ""
echo "=== Step 1: Create listing ==="
LISTING_RESP=$(curl -sf -X POST "$LISTINGS_URL/listings" \
    -H "Content-Type: application/json" \
    -d '{"sellerId":"smoke-seller","title":"Smoke Test Bike","priceCents":5000,"category":"sporting-goods","lat":47.6062,"lon":-122.3321}')
echo "Response: $LISTING_RESP"
LISTING_ID=$(echo "$LISTING_RESP" | jq -r '.id // empty')
[ -n "$LISTING_ID" ] || { echo "FAIL step 1: no id in response"; exit 1; }
echo "listingId=$LISTING_ID"

echo ""
echo "=== Step 2: Place order ==="
ORDER_RESP=$(curl -sf -X POST "$PAYMENTS_URL/orders" \
    -H "Content-Type: application/json" \
    -d "{\"listingId\":\"$LISTING_ID\",\"buyerId\":\"smoke-buyer\",\"sellerId\":\"smoke-seller\",\"amountCents\":5000}")
echo "Response: $ORDER_RESP"
ORDER_ID=$(echo "$ORDER_RESP" | jq -r '.id // empty')
ORDER_STATUS=$(echo "$ORDER_RESP" | jq -r '.status // empty')
[ -n "$ORDER_ID" ] || { echo "FAIL step 2: no id in response"; exit 1; }
[ "$ORDER_STATUS" = "HELD" ] || { echo "FAIL step 2: expected status=HELD, got '$ORDER_STATUS'"; exit 1; }
echo "orderId=$ORDER_ID status=$ORDER_STATUS"

echo ""
echo "=== Step 3: Confirm delivery ==="
CONFIRM_RESP=$(curl -sf -X POST "$PAYMENTS_URL/orders/$ORDER_ID/confirm-delivery")
echo "Response: $CONFIRM_RESP"
FINAL_STATUS=$(echo "$CONFIRM_RESP" | jq -r '.status // empty')
[ "$FINAL_STATUS" = "RELEASED" ] || { echo "FAIL step 3: expected status=RELEASED, got '$FINAL_STATUS'"; exit 1; }
echo "status=$FINAL_STATUS"

END=$(date +%s)
echo ""
echo "PASS ($((END - START))s)"
```

- [ ] **Step 2: Make script executable**

```bash
chmod +x scripts/smoke-test.sh
```

- [ ] **Step 3: Start docker-compose infra**

```bash
cd infra && docker compose up -d && cd ..
```

Expected: Postgres, Redis, OpenSearch, and Redpanda containers start. Confirm with:
```bash
docker compose -f infra/docker-compose.yml ps
```
All four containers should show `healthy` or `running`.

- [ ] **Step 4: Start listings-service and payments-service**

In two separate terminals (or background processes):

```bash
# Terminal 1
./gradlew :listings-service:run

# Terminal 2
./gradlew :payments-service:run
```

Wait for both to print `Application started` (or similar Netty startup log) before proceeding.

- [ ] **Step 5: Run the smoke test**

```bash
./scripts/smoke-test.sh
```

Expected output:
```
Waiting for listings-service ready
Waiting for payments-service ready

=== Step 1: Create listing ===
Response: {"id":"...","sellerId":"smoke-seller","title":"Smoke Test Bike",...}
listingId=<uuid>

=== Step 2: Place order ===
Response: {"id":"...","listingId":"<uuid>","buyerId":"smoke-buyer","sellerId":"smoke-seller","amountCents":5000,"status":"HELD"}
orderId=<uuid> status=HELD

=== Step 3: Confirm delivery ===
Response: {"orderId":"<uuid>","status":"RELEASED"}
status=RELEASED

PASS (Xs)
```

If Step 2 fails with a connection error, confirm payments-service is running on port 8084. If the order status comes back as something other than `HELD`, check `OrderRepository.createWithHold` — the `application { }` module must be wiring the DB connection before serving requests.

- [ ] **Step 6: Stop services and commit**

Stop both Gradle run processes (Ctrl+C). Leave docker-compose running for Task 5 if you're continuing immediately, otherwise `docker compose -f infra/docker-compose.yml down`.

```bash
git add scripts/smoke-test.sh
git commit -m "feat: add smoke-test.sh — E2E curl verification of purchase flow"
```

---

## Task 5: kind cluster deploy + smoke test

**Files:** No new files — uses existing `scripts/build-images.sh` and `scripts/deploy-kind.sh`.

**Prerequisites:**
- `kind` installed (`brew install kind`)
- `kubectl` installed (`brew install kubectl`)
- Docker running with sufficient resources (at least 4 GB RAM available)
- `jq` installed

- [ ] **Step 1: Build Docker images for all four services**

```bash
./scripts/build-images.sh
```

Expected: builds `c2c/listings-service:local`, `c2c/search-service:local`, `c2c/messaging-service:local`, `c2c/payments-service:local`. Confirm:

```bash
docker images | grep c2c
```

- [ ] **Step 2: Deploy to kind**

```bash
./scripts/deploy-kind.sh
```

Expected: creates a kind cluster named `c2c-marketplace`, loads images, applies k8s manifests. Watch pods come up:

```bash
kubectl -n c2c get pods -w
```

Wait until `listings` and `payments` pods show `Running` and `READY 1/1`. The other services can be ignored for this slice. This may take 2–3 minutes as Postgres initialises.

- [ ] **Step 3: Port-forward listings-service and payments-service**

Open two terminals and run (leave them running):

```bash
# Terminal 1
kubectl -n c2c port-forward svc/listings 8081:8081

# Terminal 2
kubectl -n c2c port-forward svc/payments 8084:8084
```

If `svc/listings` or `svc/payments` doesn't exist, check the k8s manifest names:
```bash
kubectl -n c2c get svc
```
Use the exact service name shown.

- [ ] **Step 4: Run the smoke test against kind**

```bash
./scripts/smoke-test.sh
```

Same expected output as Task 4 Step 5. The `LISTINGS_URL` and `PAYMENTS_URL` env vars remain at their defaults (`http://localhost:8081` and `http://localhost:8084`) because port-forward maps the same ports locally.

- [ ] **Step 5: Verify pods are healthy**

```bash
kubectl -n c2c get pods
```

Expected: listings and payments pods show `Running` / `1/1 READY` and low restart counts.

- [ ] **Step 6: Tear down**

```bash
kind delete cluster --name c2c-marketplace
```

- [ ] **Step 7: Final commit**

```bash
git add -u
git commit -m "chore: purchase flow vertical slice complete — all tests green, smoke-test passes on docker-compose and kind"
```

---

## Verification Checklist

Before calling this slice done:

- [ ] `./gradlew test` — BUILD SUCCESSFUL, 0 failures
- [ ] `./gradlew :payments-service:test` — `EscrowStateMachineTest` + `OrderRepositoryTest` + `PaymentsRoutesTest` all green
- [ ] `./gradlew :listings-service:test` — `ListingRepositoryTest` + `ListingRoutesTest` all green
- [ ] `./scripts/smoke-test.sh` against docker-compose — prints `PASS`
- [ ] `./scripts/smoke-test.sh` against kind (with port-forward) — prints `PASS`
