package com.patterns.saga

import io.mockk.every
import io.mockk.mockk
import io.mockk.verify
import io.temporal.client.WorkflowClient
import io.temporal.client.WorkflowOptions
import io.temporal.testing.TestWorkflowEnvironment
import io.temporal.worker.Worker
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import java.time.Duration
import java.util.UUID

/**
 * Tests for [OrderSagaWorkflowImpl] using [TestWorkflowEnvironment].
 *
 * TestWorkflowEnvironment:
 *  - Runs the workflow in a single-threaded, deterministic test environment
 *  - No Temporal server required — runs entirely in-process
 *  - Time can be fast-forwarded (not needed here but useful for timer-based workflows)
 *  - Activities are mocked using Mockk to control success/failure scenarios
 */
class OrderSagaWorkflowTest {

    private lateinit var testEnv: TestWorkflowEnvironment
    private lateinit var worker: Worker
    private lateinit var client: WorkflowClient
    private lateinit var mockActivities: OrderActivities

    private val testOrder = OrderRequest(
        orderId = "order-test-${UUID.randomUUID()}",
        customerId = "cust-test",
        items = listOf(OrderItem("SKU-TEST", quantity = 1, unitPriceCents = 1000L)),
        totalCents = 1000L,
    )

    @BeforeEach
    fun setUp() {
        testEnv = TestWorkflowEnvironment.newInstance()
        worker = testEnv.newWorker(TASK_QUEUE)
        worker.registerWorkflowImplementationTypes(OrderSagaWorkflowImpl::class.java)

        mockActivities = mockk<OrderActivities>(relaxed = false)
        worker.registerActivitiesImplementations(mockActivities)

        testEnv.start()
        client = testEnv.workflowClient
    }

    @AfterEach
    fun tearDown() {
        testEnv.close()
    }

    private fun newWorkflowStub(): OrderSagaWorkflow = client.newWorkflowStub(
        OrderSagaWorkflow::class.java,
        WorkflowOptions.newBuilder()
            .setTaskQueue(TASK_QUEUE)
            .setWorkflowId("test-saga-${UUID.randomUUID()}")
            .setWorkflowExecutionTimeout(Duration.ofSeconds(30))
            .build(),
    )

    // ─── Happy path ───────────────────────────────────────────────────────────

    @Test
    fun `successful saga completes with all three step results`() {
        every { mockActivities.chargePayment(any(), any(), any()) } returns "pay-001"
        every { mockActivities.reserveInventory(any(), any()) } returns "res-001"
        every { mockActivities.scheduleShipment(any(), any(), any()) } returns "ship-001"

        val result = newWorkflowStub().execute(testOrder)

        assertThat(result.status).isEqualTo(SagaStatus.COMPLETED)
        assertThat(result.paymentId).isEqualTo("pay-001")
        assertThat(result.reservationId).isEqualTo("res-001")
        assertThat(result.shipmentId).isEqualTo("ship-001")
        assertThat(result.failureReason).isNull()

        // Verify no compensations were called
        verify(exactly = 0) { mockActivities.refundPayment(any(), any()) }
        verify(exactly = 0) { mockActivities.releaseInventory(any(), any()) }
    }

    // ─── Failure at step 1 (chargePayment) ───────────────────────────────────

    @Test
    fun `chargePayment failure returns COMPENSATED with no compensations needed`() {
        every { mockActivities.chargePayment(any(), any(), any()) } throws
            RuntimeException("Payment gateway unavailable")

        val result = newWorkflowStub().execute(testOrder)

        assertThat(result.status).isEqualTo(SagaStatus.COMPENSATED)
        assertThat(result.failureReason).contains("Payment failed")

        // No compensations needed — payment never succeeded
        verify(exactly = 0) { mockActivities.refundPayment(any(), any()) }
        verify(exactly = 0) { mockActivities.releaseInventory(any(), any()) }
        verify(exactly = 0) { mockActivities.scheduleShipment(any(), any(), any()) }
    }

    // ─── Failure at step 2 (reserveInventory) ────────────────────────────────

    @Test
    fun `reserveInventory failure triggers refundPayment compensation`() {
        every { mockActivities.chargePayment(any(), any(), any()) } returns "pay-002"
        every { mockActivities.reserveInventory(any(), any()) } throws
            RuntimeException("Insufficient stock")
        every { mockActivities.refundPayment(any(), any()) } returns Unit

        val result = newWorkflowStub().execute(testOrder)

        assertThat(result.status).isEqualTo(SagaStatus.COMPENSATED)
        assertThat(result.failureReason).contains("Inventory reservation failed")

        // Verify compensation was called with the correct paymentId
        verify(exactly = 1) { mockActivities.refundPayment("pay-002", testOrder.orderId) }

        // Shipment should never have been called
        verify(exactly = 0) { mockActivities.scheduleShipment(any(), any(), any()) }
        // Inventory was never reserved, so no need to release it
        verify(exactly = 0) { mockActivities.releaseInventory(any(), any()) }
    }

    // ─── Failure at step 3 (scheduleShipment) ────────────────────────────────

    @Test
    fun `scheduleShipment failure triggers both releaseInventory and refundPayment compensations`() {
        every { mockActivities.chargePayment(any(), any(), any()) } returns "pay-003"
        every { mockActivities.reserveInventory(any(), any()) } returns "res-003"
        every { mockActivities.scheduleShipment(any(), any(), any()) } throws
            RuntimeException("Courier API timeout")
        every { mockActivities.releaseInventory(any(), any()) } returns Unit
        every { mockActivities.refundPayment(any(), any()) } returns Unit

        val result = newWorkflowStub().execute(testOrder)

        assertThat(result.status).isEqualTo(SagaStatus.COMPENSATED)
        assertThat(result.failureReason).contains("Shipment scheduling failed")

        // Both compensations must fire in reverse order
        verify(exactly = 1) { mockActivities.releaseInventory("res-003", testOrder.orderId) }
        verify(exactly = 1) { mockActivities.refundPayment("pay-003", testOrder.orderId) }
    }

    // ─── Compensation resilience ──────────────────────────────────────────────

    @Test
    fun `compensation failure on releaseInventory still runs refundPayment`() {
        // scheduleShipment fails; releaseInventory also fails (e.g., inventory service down).
        // refundPayment should still run — one compensation failure must not block others.
        every { mockActivities.chargePayment(any(), any(), any()) } returns "pay-004"
        every { mockActivities.reserveInventory(any(), any()) } returns "res-004"
        every { mockActivities.scheduleShipment(any(), any(), any()) } throws RuntimeException("Courier down")
        every { mockActivities.releaseInventory(any(), any()) } throws RuntimeException("Inventory service down")
        every { mockActivities.refundPayment(any(), any()) } returns Unit

        val result = newWorkflowStub().execute(testOrder)

        // Workflow still completes (as COMPENSATED) even though releaseInventory failed
        assertThat(result.status).isEqualTo(SagaStatus.COMPENSATED)

        // Both compensations were attempted
        verify(exactly = 1) { mockActivities.releaseInventory("res-004", testOrder.orderId) }
        verify(exactly = 1) { mockActivities.refundPayment("pay-004", testOrder.orderId) }
    }

    // ─── Idempotency key in activities ────────────────────────────────────────

    @Test
    fun `all three forward activities are called with correct orderId`() {
        every { mockActivities.chargePayment(any(), any(), any()) } returns "pay-idem"
        every { mockActivities.reserveInventory(any(), any()) } returns "res-idem"
        every { mockActivities.scheduleShipment(any(), any(), any()) } returns "ship-idem"

        newWorkflowStub().execute(testOrder)

        verify { mockActivities.chargePayment(testOrder.orderId, testOrder.customerId, testOrder.totalCents) }
        verify { mockActivities.reserveInventory(testOrder.orderId, testOrder.items) }
        verify { mockActivities.scheduleShipment(testOrder.orderId, testOrder.customerId, "res-idem") }
    }
}
