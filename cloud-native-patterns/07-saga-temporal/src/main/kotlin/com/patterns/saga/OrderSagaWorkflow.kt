package com.patterns.saga

import io.temporal.activity.ActivityOptions
import io.temporal.workflow.Workflow
import io.temporal.workflow.WorkflowInterface
import io.temporal.workflow.WorkflowMethod
import org.slf4j.LoggerFactory
import java.time.Duration

// ─── Data Classes ─────────────────────────────────────────────────────────────

data class OrderRequest(
    val orderId: String,
    val customerId: String,
    val items: List<OrderItem>,
    val totalCents: Long,
)

data class OrderItem(
    val sku: String,
    val quantity: Int,
    val unitPriceCents: Long,
)

data class OrderResult(
    val orderId: String,
    val status: SagaStatus,
    val paymentId: String? = null,
    val reservationId: String? = null,
    val shipmentId: String? = null,
    val failureReason: String? = null,
) {
    companion object {
        fun success(orderId: String, paymentId: String, reservationId: String, shipmentId: String) =
            OrderResult(orderId, SagaStatus.COMPLETED, paymentId, reservationId, shipmentId)

        fun failed(orderId: String, reason: String) =
            OrderResult(orderId, SagaStatus.COMPENSATED, failureReason = reason)
    }
}

enum class SagaStatus { COMPLETED, COMPENSATED }

// ─── Workflow Interface ───────────────────────────────────────────────────────

/**
 * Orchestrated saga with three forward steps and compensating actions.
 *
 * Step sequence:
 *   chargePayment → reserveInventory → scheduleShipment
 *
 * Compensation on failure:
 *   scheduleShipment fails → releaseInventory + refundPayment
 *   reserveInventory fails → refundPayment
 *   chargePayment fails   → nothing to compensate
 *
 * The compensation strategy must be designed FIRST — before writing the happy path.
 * This is the most critical discipline in saga design.
 */
@WorkflowInterface
interface OrderSagaWorkflow {
    @WorkflowMethod
    fun execute(order: OrderRequest): OrderResult
}

// ─── Workflow Implementation ──────────────────────────────────────────────────

class OrderSagaWorkflowImpl : OrderSagaWorkflow {

    // Logger must use Workflow.getLogger() — not LoggerFactory — to ensure replay-safe logging.
    private val log = Workflow.getLogger(OrderSagaWorkflowImpl::class.java)

    // Activity stub with start-to-close timeout.
    // Temporal retries activities automatically on failure up to scheduleToCloseTimeout.
    private val activities: OrderActivities = Workflow.newActivityStub(
        OrderActivities::class.java,
        ActivityOptions.newBuilder()
            .setStartToCloseTimeout(Duration.ofSeconds(10))
            .setScheduleToCloseTimeout(Duration.ofMinutes(2))
            .build(),
    )

    override fun execute(order: OrderRequest): OrderResult {
        log.info("OrderSaga starting for orderId={}", order.orderId)

        // Forward step 1: charge payment
        val paymentId = try {
            activities.chargePayment(order.orderId, order.customerId, order.totalCents)
        } catch (e: Exception) {
            log.error("chargePayment failed — no compensation needed: {}", e.message)
            return OrderResult.failed(order.orderId, "Payment failed: ${e.message}")
        }

        // Forward step 2: reserve inventory
        val reservationId = try {
            activities.reserveInventory(order.orderId, order.items)
        } catch (e: Exception) {
            log.warn("reserveInventory failed — compensating: refundPayment paymentId={}", paymentId)
            // Compensate step 1
            runCompensation("refundPayment") {
                activities.refundPayment(paymentId, order.orderId)
            }
            return OrderResult.failed(order.orderId, "Inventory reservation failed: ${e.message}")
        }

        // Forward step 3: schedule shipment
        val shipmentId = try {
            activities.scheduleShipment(order.orderId, order.customerId, reservationId)
        } catch (e: Exception) {
            log.warn(
                "scheduleShipment failed — compensating: releaseInventory reservationId={}, refundPayment paymentId={}",
                reservationId,
                paymentId,
            )
            // Compensate steps 2 and 1 (in reverse order)
            runCompensation("releaseInventory") {
                activities.releaseInventory(reservationId, order.orderId)
            }
            runCompensation("refundPayment") {
                activities.refundPayment(paymentId, order.orderId)
            }
            return OrderResult.failed(order.orderId, "Shipment scheduling failed: ${e.message}")
        }

        log.info(
            "OrderSaga completed successfully orderId={} paymentId={} reservationId={} shipmentId={}",
            order.orderId,
            paymentId,
            reservationId,
            shipmentId,
        )

        return OrderResult.success(order.orderId, paymentId, reservationId, shipmentId)
    }

    /**
     * Runs a compensating action, swallowing exceptions so that one compensation
     * failure doesn't prevent subsequent compensations from running.
     *
     * In production, failed compensations should be logged to an alert channel —
     * they require manual intervention or a dead-letter queue.
     */
    private fun runCompensation(name: String, action: () -> Unit) {
        try {
            action()
            log.info("Compensation '{}' succeeded", name)
        } catch (e: Exception) {
            // Swallow and log — don't let one compensation failure block others.
            // Alert: this requires manual intervention or a separate compensation workflow.
            log.error("Compensation '{}' FAILED — requires manual review: {}", name, e.message)
        }
    }
}
