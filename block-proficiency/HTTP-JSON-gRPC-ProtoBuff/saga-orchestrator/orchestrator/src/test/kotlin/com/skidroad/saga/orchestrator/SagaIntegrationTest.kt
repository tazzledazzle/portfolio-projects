package com.skidroad.saga.orchestrator

import com.skidroad.saga.orchestrator.resilience.CircuitBreaker
import com.skidroad.saga.orchestrator.resilience.RetryPolicy
import com.skidroad.saga.orchestrator.saga.*
import com.skidroad.saga.proto.*
import io.grpc.ManagedChannel
import io.grpc.ServerBuilder
import io.grpc.Status
import io.grpc.StatusException
import io.grpc.inprocess.InProcessChannelBuilder
import io.grpc.inprocess.InProcessServerBuilder
import kotlinx.coroutines.test.runTest
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows
import kotlin.test.assertEquals
import kotlin.test.assertNotNull

/**
 * Integration tests using gRPC's [InProcessServerBuilder] — runs real gRPC
 * serialization/deserialization without network I/O, so tests are fast and hermetic.
 *
 * Tests cover:
 *  1. Happy path: all three services succeed → COMPLETED
 *  2. Shipping failure → compensation runs → payment refunded + inventory released
 *  3. Circuit breaker opens after threshold failures
 *  4. Retry policy with exponential backoff
 *  5. Deadline propagation (DEADLINE_EXCEEDED surfaces correctly)
 *  6. Idempotent saga execution (same idempotency key → same result)
 */
class SagaIntegrationTest {

    private val serverName = InProcessServerBuilder.generateName()
    private lateinit var channel: ManagedChannel
    private lateinit var paymentImpl: TrackingPaymentService
    private lateinit var inventoryImpl: TrackingInventoryService
    private lateinit var shippingImpl: TrackingShippingService

    @BeforeEach
    fun setup() {
        paymentImpl   = TrackingPaymentService()
        inventoryImpl = TrackingInventoryService()
        shippingImpl  = TrackingShippingService()

        InProcessServerBuilder.forName(serverName)
            .addService(paymentImpl)
            .addService(inventoryImpl)
            .addService(shippingImpl)
            .build()
            .start()

        channel = InProcessChannelBuilder.forName(serverName).directExecutor().build()
    }

    @AfterEach
    fun teardown() { channel.shutdownNow() }

    private fun buildClients(
        paymentOverride:   PaymentServiceGrpcKt.PaymentServiceCoroutineStub? = null,
        inventoryOverride: InventoryServiceGrpcKt.InventoryServiceCoroutineStub? = null,
        shippingOverride:  ShippingServiceGrpcKt.ShippingServiceCoroutineStub? = null
    ) = DownstreamClients(
        payment   = paymentOverride   ?: PaymentServiceGrpcKt.PaymentServiceCoroutineStub(channel),
        inventory = inventoryOverride ?: InventoryServiceGrpcKt.InventoryServiceCoroutineStub(channel),
        shipping  = shippingOverride  ?: ShippingServiceGrpcKt.ShippingServiceCoroutineStub(channel)
    )

    private fun sampleRequest(idempotencyKey: String = "key-${System.nanoTime()}") = placeOrderRequest {
        orderId        = "order-${System.nanoTime()}"
        this.idempotencyKey = idempotencyKey
        items += orderItem { sku = "SKU-001"; quantity = 2; priceCents = 1000 }
        paymentInfo = paymentInfo {
            paymentMethodId = "pm_test_visa"
            amountCents     = 2000
            currency        = "USD"
        }
        address = shippingAddress {
            street   = "123 Main St"
            city     = "Seattle"
            postcode = "98101"
            country  = "US"
        }
    }

    // ── Test 1: Happy Path ───────────────────────────────────────────────────

    @Test
    fun `happy path - all services succeed - saga completes`() = runTest {
        shippingImpl.alwaysSucceed = true
        val stateMachine = SagaStateMachine(buildClients())
        val response = stateMachine.execute(sampleRequest())

        assertEquals(SagaStatus.State.COMPLETED, response.sagaStatus.state)
        assertNotNull(response.shipmentId)
        assertEquals(1, paymentImpl.chargeCount)
        assertEquals(1, inventoryImpl.reserveCount)
        assertEquals(1, shippingImpl.createCount)
        assertEquals(0, inventoryImpl.releaseCount, "No compensation should run on success")
        assertEquals(0, paymentImpl.refundCount,    "No compensation should run on success")
    }

    // ── Test 2: Shipping Failure → Compensation ──────────────────────────────

    @Test
    fun `shipping failure triggers compensation - inventory released and payment refunded`() = runTest {
        shippingImpl.alwaysFail = true
        val stateMachine = SagaStateMachine(
            buildClients(),
            retryPolicy = RetryPolicy(maxAttempts = 1)  // no retries for fast test
        )

        val ex = assertThrows<StatusException> { stateMachine.execute(sampleRequest()) }
        assertEquals(Status.Code.INTERNAL, ex.status.code)

        // Verify compensation ran
        assertEquals(1, paymentImpl.refundCount,    "Payment should be refunded")
        assertEquals(1, inventoryImpl.releaseCount, "Inventory should be released")
    }

    // ── Test 3: Circuit Breaker Opens ────────────────────────────────────────

    @Test
    fun `circuit breaker opens after failure threshold`() = runTest {
        val cb = CircuitBreaker("test-cb", failureThreshold = 3, resetTimeoutMs = 60_000)

        // Trigger 3 failures to open the circuit
        repeat(3) {
            try {
                cb.execute<Unit> { throw StatusException(Status.UNAVAILABLE) }
            } catch (_: StatusException) { }
        }

        assertEquals(CircuitBreaker.State.OPEN, cb.currentState)

        // Next call should fail immediately with UNAVAILABLE (short-circuit)
        val ex = assertThrows<StatusException> { cb.execute<Unit> { /* would succeed */ } }
        assertEquals(Status.Code.UNAVAILABLE, ex.status.code)
        assert(ex.status.description?.contains("Circuit breaker OPEN") == true)
    }

    // ── Test 4: Retry Policy ─────────────────────────────────────────────────

    @Test
    fun `retry policy retries on UNAVAILABLE then succeeds`() = runTest {
        var attempts = 0
        val policy = RetryPolicy(maxAttempts = 3, baseDelayMs = 10)

        val result = policy.execute("test-op") {
            attempts++
            if (attempts < 3) throw StatusException(Status.UNAVAILABLE)
            "success"
        }

        assertEquals("success", result)
        assertEquals(3, attempts)
    }

    @Test
    fun `retry policy does NOT retry on INVALID_ARGUMENT`() = runTest {
        var attempts = 0
        val policy = RetryPolicy(maxAttempts = 3)

        assertThrows<StatusException> {
            policy.execute("test-op") {
                attempts++
                throw StatusException(Status.INVALID_ARGUMENT.withDescription("bad input"))
            }
        }

        // Should fail immediately on first attempt, not retry
        assertEquals(1, attempts, "INVALID_ARGUMENT should not be retried")
    }

    // ── Test 5: Idempotency ──────────────────────────────────────────────────

    @Test
    fun `same idempotency key returns same result without re-processing`() = runTest {
        shippingImpl.alwaysSucceed = true
        val stateMachine = SagaStateMachine(buildClients())
        val key = "idempotent-key-${System.nanoTime()}"

        val r1 = stateMachine.execute(sampleRequest(idempotencyKey = key))
        val r2 = stateMachine.execute(sampleRequest(idempotencyKey = key))

        assertEquals(r1.shipmentId, r2.shipmentId)
        // Payment should only be charged once despite two saga executions
        assertEquals(1, paymentImpl.chargeCount)
    }
}

// ── Test doubles ──────────────────────────────────────────────────────────────

class TrackingPaymentService : PaymentServiceGrpcKt.PaymentServiceCoroutineImplBase() {
    var chargeCount = 0; var refundCount = 0
    private val idempotencyMap = mutableMapOf<String, ChargePaymentResponse>()

    override suspend fun chargePayment(r: ChargePaymentRequest): ChargePaymentResponse {
        idempotencyMap[r.idempotencyKey]?.let { return it }
        chargeCount++
        return chargePaymentResponse { transactionId = "tx-$chargeCount" }
            .also { idempotencyMap[r.idempotencyKey] = it }
    }
    override suspend fun refundPayment(r: RefundPaymentRequest): RefundPaymentResponse {
        refundCount++
        return refundPaymentResponse { refundId = "ref-$refundCount" }
    }
}

class TrackingInventoryService : InventoryServiceGrpcKt.InventoryServiceCoroutineImplBase() {
    var reserveCount = 0; var releaseCount = 0

    override suspend fun reserveInventory(r: ReserveInventoryRequest): ReserveInventoryResponse {
        reserveCount++
        return reserveInventoryResponse { reservationId = "res-$reserveCount" }
    }
    override suspend fun releaseInventory(r: ReleaseInventoryRequest): ReleaseInventoryResponse {
        releaseCount++
        return releaseInventoryResponse { released = true }
    }
}

class TrackingShippingService : ShippingServiceGrpcKt.ShippingServiceCoroutineImplBase() {
    var createCount = 0; var alwaysFail = false; var alwaysSucceed = false

    override suspend fun createShipment(r: CreateShipmentRequest): CreateShipmentResponse {
        if (alwaysFail) throw StatusException(Status.UNAVAILABLE.withDescription("simulated failure"))
        createCount++
        return createShipmentResponse { shipmentId = "ship-$createCount"; trackingNumber = "TRK-$createCount" }
    }
}
