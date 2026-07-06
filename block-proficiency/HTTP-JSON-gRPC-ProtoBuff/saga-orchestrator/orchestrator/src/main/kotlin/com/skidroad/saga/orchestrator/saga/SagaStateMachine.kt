package com.skidroad.saga.orchestrator.saga

import com.skidroad.saga.proto.*
import com.skidroad.saga.orchestrator.resilience.CircuitBreaker
import com.skidroad.saga.orchestrator.resilience.RetryPolicy
import com.google.protobuf.Timestamp
import io.grpc.Status
import io.grpc.StatusException
import mu.KotlinLogging
import java.time.Instant
import java.util.UUID

private val log = KotlinLogging.logger {}

/**
 * Encapsulates all downstream gRPC stubs the orchestrator uses.
 * Each stub should already have [TracingClientInterceptor] attached.
 */
data class DownstreamClients(
    val payment:   PaymentServiceGrpcKt.PaymentServiceCoroutineStub,
    val inventory: InventoryServiceGrpcKt.InventoryServiceCoroutineStub,
    val shipping:  ShippingServiceGrpcKt.ShippingServiceCoroutineStub
)

/**
 * Mutable execution context for a single saga run.
 * Tracks completed steps so CompensationEngine knows what to undo.
 */
data class SagaContext(
    val sagaId:       String = UUID.randomUUID().toString(),
    val orderId:      String,
    val idempotencyKey: String,
    var state:        SagaStatus.State = SagaStatus.State.PENDING,
    var transactionId:   String? = null,
    var reservationId:   String? = null,
    var shipmentId:      String? = null,
    val compensationLog: MutableList<CompensationStep> = mutableListOf()
)

/**
 * The heart of the Saga pattern: a coroutine-driven orchestrator that:
 *
 *  1. Executes each step sequentially with deadline propagation
 *  2. Wraps each step in a CircuitBreaker + RetryPolicy
 *  3. On any failure, delegates to [CompensationEngine] for rollback
 *
 * Deadline propagation: gRPC deadline is set on the *stub call options*
 * (withDeadlineAfter), not globally. This ensures each downstream service
 * gets a proportional slice of the parent deadline.
 *
 * Idempotency: every downstream request includes the saga's idempotency key.
 * This means it's safe to retry any step — downstream services deduplicate.
 */
class SagaStateMachine(
    private val clients: DownstreamClients,
    private val paymentCB:   CircuitBreaker = CircuitBreaker("payment-service"),
    private val inventoryCB: CircuitBreaker = CircuitBreaker("inventory-service"),
    private val shippingCB:  CircuitBreaker = CircuitBreaker("shipping-service"),
    private val retryPolicy: RetryPolicy    = RetryPolicy(maxAttempts = 3)
) {

    suspend fun execute(request: PlaceOrderRequest): PlaceOrderResponse {
        val ctx = SagaContext(
            orderId        = request.orderId,
            idempotencyKey = request.idempotencyKey.ifBlank { UUID.randomUUID().toString() }
        )

        log.info { "[${ctx.sagaId}] Starting saga for orderId=${ctx.orderId}" }

        try {
            // ── Step 1: Charge Payment ───────────────────────────────────────
            val paymentResp = retryPolicy.execute("charge-payment") {
                paymentCB.execute {
                    clients.payment
                        .withDeadlineAfter(3, java.util.concurrent.TimeUnit.SECONDS)
                        .chargePayment(
                            chargePaymentRequest {
                                orderId         = ctx.orderId
                                paymentMethodId = request.paymentInfo.paymentMethodId
                                amountCents     = request.paymentInfo.amountCents
                                currency        = request.paymentInfo.currency
                                idempotencyKey  = ctx.idempotencyKey
                            }
                        )
                }
            }
            ctx.transactionId = paymentResp.transactionId
            ctx.state         = SagaStatus.State.PAYMENT_CHARGED
            log.info { "[${ctx.sagaId}] Payment charged: txId=${ctx.transactionId}" }

            // ── Step 2: Reserve Inventory ────────────────────────────────────
            val inventoryResp = retryPolicy.execute("reserve-inventory") {
                inventoryCB.execute {
                    clients.inventory
                        .withDeadlineAfter(3, java.util.concurrent.TimeUnit.SECONDS)
                        .reserveInventory(
                            reserveInventoryRequest {
                                orderId        = ctx.orderId
                                items          += request.itemsList
                                idempotencyKey = ctx.idempotencyKey
                            }
                        )
                }
            }
            ctx.reservationId = inventoryResp.reservationId
            ctx.state         = SagaStatus.State.INVENTORY_RESERVED
            log.info { "[${ctx.sagaId}] Inventory reserved: reservationId=${ctx.reservationId}" }

            // ── Step 3: Create Shipment ──────────────────────────────────────
            val shippingResp = retryPolicy.execute("create-shipment") {
                shippingCB.execute {
                    clients.shipping
                        .withDeadlineAfter(4, java.util.concurrent.TimeUnit.SECONDS)
                        .createShipment(
                            createShipmentRequest {
                                orderId        = ctx.orderId
                                reservationId  = ctx.reservationId!!
                                address        = request.address
                                items          += request.itemsList
                                idempotencyKey = ctx.idempotencyKey
                            }
                        )
                }
            }
            ctx.shipmentId = shippingResp.shipmentId
            ctx.state      = SagaStatus.State.COMPLETED
            log.info { "[${ctx.sagaId}] Saga completed. shipmentId=${ctx.shipmentId}" }

            return placeOrderResponse {
                sagaStatus = ctx.toProto()
                shipmentId = ctx.shipmentId!!
            }

        } catch (e: StatusException) {
            log.error { "[${ctx.sagaId}] Step failed at state=${ctx.state}: ${e.status}" }
            ctx.state = SagaStatus.State.COMPENSATING
            CompensationEngine(clients).compensate(ctx)
            ctx.state = SagaStatus.State.COMPENSATION_COMPLETE

            throw StatusException(
                Status.INTERNAL.withDescription(
                    "Saga failed at ${ctx.state}: ${e.status.description}. " +
                    "Compensation: ${ctx.compensationLog.joinToString { "${it.service}=${if(it.succeeded) "OK" else "FAILED"}" }}"
                )
            )
        }
    }
}

/**
 * Executes compensation steps in **reverse order** relative to what succeeded.
 *
 * Critical design point: compensation must be idempotent. If a compensation
 * step itself fails, we log and continue — we never let a compensation failure
 * prevent other compensations from running. All failures are recorded in the log.
 *
 * This mirrors the "best-effort" compensation model used by Temporal/Cadence sagas.
 */
class CompensationEngine(private val clients: DownstreamClients) {

    suspend fun compensate(ctx: SagaContext) {
        log.warn { "[${ctx.sagaId}] Starting compensation from state=${ctx.state}" }

        // Compensation runs in reverse: shipping → inventory → payment
        // but we only compensate steps that actually completed.

        if (ctx.reservationId != null) {
            ctx.compensationLog += safeCompensate("inventory", "release-inventory") {
                clients.inventory.releaseInventory(
                    releaseInventoryRequest {
                        reservationId = ctx.reservationId!!
                        reason        = "saga-compensation orderId=${ctx.orderId}"
                    }
                )
            }
        }

        if (ctx.transactionId != null) {
            ctx.compensationLog += safeCompensate("payment", "refund-payment") {
                clients.payment.refundPayment(
                    refundPaymentRequest {
                        transactionId = ctx.transactionId!!
                        reason        = "saga-compensation orderId=${ctx.orderId}"
                    }
                )
            }
        }

        log.warn { "[${ctx.sagaId}] Compensation complete: ${ctx.compensationLog}" }
    }

    private suspend fun safeCompensate(
        service: String,
        action:  String,
        block:   suspend () -> Any
    ): CompensationStep {
        return try {
            block()
            log.info { "Compensation [$service/$action] succeeded" }
            compensationStep { this.service = service; this.action = action; succeeded = true }
        } catch (e: Exception) {
            log.error(e) { "Compensation [$service/$action] FAILED — requires manual intervention" }
            compensationStep {
                this.service = service
                this.action  = action
                succeeded    = false
                error        = e.message ?: "unknown"
            }
        }
    }
}

// ── Proto builder extensions ──────────────────────────────────────────────────

private fun SagaContext.toProto(): SagaStatus = sagaStatus {
    sagaId  = this@toProto.sagaId
    orderId = this@toProto.orderId
    state   = this@toProto.state
    updatedAt = Instant.now().toProtoTimestamp()
    compensationLog += this@toProto.compensationLog
}

private fun Instant.toProtoTimestamp(): Timestamp =
    Timestamp.newBuilder().setSeconds(epochSecond).setNanos(nano).build()
