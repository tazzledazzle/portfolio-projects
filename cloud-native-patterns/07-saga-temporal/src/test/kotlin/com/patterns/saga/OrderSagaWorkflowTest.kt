package com.patterns.saga

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
import java.util.concurrent.atomic.AtomicInteger

/**
 * Tests for [OrderSagaWorkflowImpl] using [TestWorkflowEnvironment].
 *
 * MockK cannot be used directly with Temporal activity interfaces — MockK's ByteBuddy
 * proxy inherits @ActivityMethod annotations onto the concrete class, which Temporal
 * rejects. Use a manually written stub ([StubActivities]) instead.
 *
 * StubActivities is a plain class (no @ActivityInterface/@ActivityMethod) registered
 * as an implementation of [OrderActivities]. It records calls and can be configured
 * to throw on specific methods — same expressiveness as a mock, without the annotation
 * inheritance problem.
 */
class OrderSagaWorkflowTest {

    private lateinit var testEnv: TestWorkflowEnvironment
    private lateinit var worker: Worker
    private lateinit var client: WorkflowClient
    private lateinit var stub: StubActivities

    private val testOrder = OrderRequest(
        orderId = "order-test-${UUID.randomUUID()}",
        customerId = "cust-test",
        items = listOf(OrderItem("SKU-TEST", quantity = 1, unitPriceCents = 1000L)),
        totalCents = 1000L,
    )

    @BeforeEach
    fun setUp() {
        stub = StubActivities()
        testEnv = TestWorkflowEnvironment.newInstance()
        worker = testEnv.newWorker(TASK_QUEUE)
        worker.registerWorkflowImplementationTypes(OrderSagaWorkflowImpl::class.java)
        worker.registerActivitiesImplementations(stub)
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
        stub.chargePaymentResult = "pay-001"
        stub.reserveInventoryResult = "res-001"
        stub.scheduleShipmentResult = "ship-001"

        val result = newWorkflowStub().execute(testOrder)

        assertThat(result.status).isEqualTo(SagaStatus.COMPLETED)
        assertThat(result.paymentId).isEqualTo("pay-001")
        assertThat(result.reservationId).isEqualTo("res-001")
        assertThat(result.shipmentId).isEqualTo("ship-001")
        assertThat(result.failureReason).isNull()
        assertThat(stub.refundCallCount.get()).isEqualTo(0)
        assertThat(stub.releaseCallCount.get()).isEqualTo(0)
    }

    // ─── Failure at step 1 (chargePayment) ───────────────────────────────────

    @Test
    fun `chargePayment failure returns COMPENSATED with no compensations needed`() {
        stub.chargePaymentError = RuntimeException("Payment gateway unavailable")

        val result = newWorkflowStub().execute(testOrder)

        assertThat(result.status).isEqualTo(SagaStatus.COMPENSATED)
        assertThat(result.failureReason).contains("Payment failed")
        assertThat(stub.refundCallCount.get()).isEqualTo(0)
        assertThat(stub.releaseCallCount.get()).isEqualTo(0)
        assertThat(stub.scheduleCallCount.get()).isEqualTo(0)
    }

    // ─── Failure at step 2 (reserveInventory) ────────────────────────────────

    @Test
    fun `reserveInventory failure triggers refundPayment compensation`() {
        stub.chargePaymentResult = "pay-002"
        stub.reserveInventoryError = RuntimeException("Insufficient stock")

        val result = newWorkflowStub().execute(testOrder)

        assertThat(result.status).isEqualTo(SagaStatus.COMPENSATED)
        assertThat(result.failureReason).contains("Inventory reservation failed")
        assertThat(stub.refundCallCount.get()).isEqualTo(1)
        assertThat(stub.refundPaymentIdCapture).isEqualTo("pay-002")
        assertThat(stub.scheduleCallCount.get()).isEqualTo(0)
        assertThat(stub.releaseCallCount.get()).isEqualTo(0)
    }

    // ─── Failure at step 3 (scheduleShipment) ────────────────────────────────

    @Test
    fun `scheduleShipment failure triggers both releaseInventory and refundPayment compensations`() {
        stub.chargePaymentResult = "pay-003"
        stub.reserveInventoryResult = "res-003"
        stub.scheduleShipmentError = RuntimeException("Courier API timeout")

        val result = newWorkflowStub().execute(testOrder)

        assertThat(result.status).isEqualTo(SagaStatus.COMPENSATED)
        assertThat(result.failureReason).contains("Shipment scheduling failed")
        assertThat(stub.releaseCallCount.get()).isEqualTo(1)
        assertThat(stub.releaseReservationIdCapture).isEqualTo("res-003")
        assertThat(stub.refundCallCount.get()).isEqualTo(1)
        assertThat(stub.refundPaymentIdCapture).isEqualTo("pay-003")
    }

    // ─── Compensation resilience ──────────────────────────────────────────────

    @Test
    fun `compensation failure on releaseInventory still runs refundPayment`() {
        stub.chargePaymentResult = "pay-004"
        stub.reserveInventoryResult = "res-004"
        stub.scheduleShipmentError = RuntimeException("Courier down")
        stub.releaseInventoryError = RuntimeException("Inventory service down")

        val result = newWorkflowStub().execute(testOrder)

        assertThat(result.status).isEqualTo(SagaStatus.COMPENSATED)
        assertThat(stub.releaseCallCount.get()).isEqualTo(1)
        assertThat(stub.refundCallCount.get()).isEqualTo(1)
    }

    // ─── Activity call verification ───────────────────────────────────────────

    @Test
    fun `all three forward activities are called with correct orderId`() {
        stub.chargePaymentResult = "pay-idem"
        stub.reserveInventoryResult = "res-idem"
        stub.scheduleShipmentResult = "ship-idem"

        newWorkflowStub().execute(testOrder)

        assertThat(stub.chargeOrderIdCapture).isEqualTo(testOrder.orderId)
        assertThat(stub.reserveOrderIdCapture).isEqualTo(testOrder.orderId)
        assertThat(stub.scheduleOrderIdCapture).isEqualTo(testOrder.orderId)
        assertThat(stub.scheduleReservationIdCapture).isEqualTo("res-idem")
    }
}

/**
 * Configurable activity stub — implements [OrderActivities] as a plain class.
 *
 * Must NOT have @ActivityInterface or @ActivityMethod annotations (those belong
 * on the interface only). Temporal will match this to [OrderActivities] via
 * the interface it implements.
 */
class StubActivities : OrderActivities {
    // Configurable return values / errors
    var chargePaymentResult: String = "pay-stub"
    var chargePaymentError: RuntimeException? = null

    var reserveInventoryResult: String = "res-stub"
    var reserveInventoryError: RuntimeException? = null

    var scheduleShipmentResult: String = "ship-stub"
    var scheduleShipmentError: RuntimeException? = null

    var releaseInventoryError: RuntimeException? = null
    var refundPaymentError: RuntimeException? = null

    // Call tracking
    val chargeCallCount = AtomicInteger(0)
    val reserveCallCount = AtomicInteger(0)
    val scheduleCallCount = AtomicInteger(0)
    val releaseCallCount = AtomicInteger(0)
    val refundCallCount = AtomicInteger(0)

    // Argument captures (last call)
    var chargeOrderIdCapture: String? = null
    var reserveOrderIdCapture: String? = null
    var scheduleOrderIdCapture: String? = null
    var scheduleReservationIdCapture: String? = null
    var releaseReservationIdCapture: String? = null
    var refundPaymentIdCapture: String? = null

    override fun chargePayment(orderId: String, customerId: String, amountCents: Long): String {
        chargeCallCount.incrementAndGet()
        chargeOrderIdCapture = orderId
        chargePaymentError?.let { throw it }
        return chargePaymentResult
    }

    override fun reserveInventory(orderId: String, items: List<OrderItem>): String {
        reserveCallCount.incrementAndGet()
        reserveOrderIdCapture = orderId
        reserveInventoryError?.let { throw it }
        return reserveInventoryResult
    }

    override fun scheduleShipment(orderId: String, customerId: String, reservationId: String): String {
        scheduleCallCount.incrementAndGet()
        scheduleOrderIdCapture = orderId
        scheduleReservationIdCapture = reservationId
        scheduleShipmentError?.let { throw it }
        return scheduleShipmentResult
    }

    override fun refundPayment(paymentId: String, orderId: String) {
        refundCallCount.incrementAndGet()
        refundPaymentIdCapture = paymentId
        refundPaymentError?.let { throw it }
    }

    override fun releaseInventory(reservationId: String, orderId: String) {
        releaseCallCount.incrementAndGet()
        releaseReservationIdCapture = reservationId
        releaseInventoryError?.let { throw it }
    }
}
