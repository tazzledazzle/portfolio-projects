package com.patterns.saga

import io.temporal.activity.Activity
import io.temporal.activity.ActivityInterface
import io.temporal.activity.ActivityMethod
import org.slf4j.LoggerFactory
import java.util.UUID

// ─── Activity Interface ───────────────────────────────────────────────────────

/**
 * Activities are the "leaf nodes" of a saga — the actual service calls.
 *
 * Each activity must be idempotent: Temporal retries activities automatically
 * on worker failure. The idempotency key (derived from workflowId) lets downstream
 * services deduplicate retried calls.
 *
 * Forward activities (charge, reserve, schedule) mutate state.
 * Compensating activities (refund, release) undo that mutation.
 */
@ActivityInterface
interface OrderActivities {
    @ActivityMethod fun chargePayment(orderId: String, customerId: String, amountCents: Long): String
    @ActivityMethod fun reserveInventory(orderId: String, items: List<OrderItem>): String
    @ActivityMethod fun scheduleShipment(orderId: String, customerId: String, reservationId: String): String

    // Compensating actions — called on failure in reverse order
    @ActivityMethod fun refundPayment(paymentId: String, orderId: String)
    @ActivityMethod fun releaseInventory(reservationId: String, orderId: String)
}

// ─── Activity Implementation ──────────────────────────────────────────────────

/**
 * Production-like activity implementation with:
 *  - Idempotency key derived from workflowId (safe to retry)
 *  - Heartbeating for long-running activities
 *  - Configurable failure injection for testing saga compensation
 */
class OrderActivitiesImpl(
    private val failOnActivity: String? = null,  // Name of activity to simulate failure on
) : OrderActivities {

    private val log = LoggerFactory.getLogger(OrderActivitiesImpl::class.java)

    override fun chargePayment(orderId: String, customerId: String, amountCents: Long): String {
        val ctx = Activity.getExecutionContext()
        // Idempotency key = workflowId + activityType ensures deduplication on retry.
        // The payment service uses this key to prevent double-charging.
        val idempotencyKey = "${ctx.info.workflowId}:chargePayment"

        log.info(
            "[chargePayment] orderId={} customerId={} amount={}¢ idempotencyKey={}",
            orderId,
            customerId,
            amountCents,
            idempotencyKey,
        )

        // Heartbeat for long-running activities (payment processing can be slow)
        ctx.heartbeat("initiating charge")

        if (failOnActivity == "chargePayment") {
            throw RuntimeException("Simulated payment gateway failure for orderId=$orderId")
        }

        // Simulate payment processing latency
        Thread.sleep(50)

        val paymentId = "pay-${UUID.randomUUID()}"
        log.info("[chargePayment] SUCCESS paymentId={} orderId={}", paymentId, orderId)
        return paymentId
    }

    override fun reserveInventory(orderId: String, items: List<OrderItem>): String {
        val ctx = Activity.getExecutionContext()
        val idempotencyKey = "${ctx.info.workflowId}:reserveInventory"

        log.info(
            "[reserveInventory] orderId={} itemCount={} idempotencyKey={}",
            orderId,
            items.size,
            idempotencyKey,
        )

        ctx.heartbeat("checking stock levels")

        if (failOnActivity == "reserveInventory") {
            throw RuntimeException("Simulated inventory shortage for orderId=$orderId")
        }

        Thread.sleep(30)

        val reservationId = "res-${UUID.randomUUID()}"
        log.info("[reserveInventory] SUCCESS reservationId={} orderId={}", reservationId, orderId)
        return reservationId
    }

    override fun scheduleShipment(orderId: String, customerId: String, reservationId: String): String {
        val ctx = Activity.getExecutionContext()
        val idempotencyKey = "${ctx.info.workflowId}:scheduleShipment"

        log.info(
            "[scheduleShipment] orderId={} reservationId={} idempotencyKey={}",
            orderId,
            reservationId,
            idempotencyKey,
        )

        ctx.heartbeat("contacting courier API")

        if (failOnActivity == "scheduleShipment") {
            throw RuntimeException("Simulated courier API unavailable for orderId=$orderId")
        }

        Thread.sleep(40)

        val shipmentId = "ship-${UUID.randomUUID()}"
        log.info("[scheduleShipment] SUCCESS shipmentId={} orderId={}", shipmentId, orderId)
        return shipmentId
    }

    override fun refundPayment(paymentId: String, orderId: String) {
        val ctx = Activity.getExecutionContext()
        val idempotencyKey = "${ctx.info.workflowId}:refundPayment"

        log.info(
            "[refundPayment] COMPENSATION paymentId={} orderId={} idempotencyKey={}",
            paymentId,
            orderId,
            idempotencyKey,
        )

        // Compensating actions must also be idempotent — the refund service uses
        // idempotencyKey to avoid double-refunding on retry.
        Thread.sleep(30)
        log.info("[refundPayment] COMPENSATION SUCCESS paymentId={}", paymentId)
    }

    override fun releaseInventory(reservationId: String, orderId: String) {
        val ctx = Activity.getExecutionContext()
        val idempotencyKey = "${ctx.info.workflowId}:releaseInventory"

        log.info(
            "[releaseInventory] COMPENSATION reservationId={} orderId={} idempotencyKey={}",
            reservationId,
            orderId,
            idempotencyKey,
        )

        Thread.sleep(20)
        log.info("[releaseInventory] COMPENSATION SUCCESS reservationId={}", reservationId)
    }
}
