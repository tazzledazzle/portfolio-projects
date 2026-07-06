package com.skidroad.saga

import com.skidroad.saga.proto.*
import com.google.protobuf.Timestamp
import io.grpc.ServerBuilder
import io.grpc.Status
import io.grpc.StatusException
import mu.KotlinLogging
import java.time.Instant
import java.util.UUID
import java.util.concurrent.ConcurrentHashMap

private val log = KotlinLogging.logger {}

/**
 * Payment Service gRPC implementation.
 *
 * Demonstrates proper gRPC status code usage:
 *  - INVALID_ARGUMENT: caller's fault, don't retry
 *  - NOT_FOUND:        resource doesn't exist
 *  - ALREADY_EXISTS:   idempotency key collision (safe to return existing result)
 *  - INTERNAL:         server-side failure, may retry
 *  - UNAVAILABLE:      transient, retry with backoff (triggers circuit breaker)
 *
 * Idempotency: requests with the same idempotency_key return the stored result
 * without re-processing. This makes saga compensation safe to retry.
 */
class PaymentGrpcService : PaymentServiceGrpcKt.PaymentServiceCoroutineImplBase() {

    // Simulates a payment ledger + idempotency store
    private val transactions  = ConcurrentHashMap<String, ChargePaymentResponse>()  // txId → response
    private val idempotencyMap = ConcurrentHashMap<String, ChargePaymentResponse>() // key → response
    private val refunds       = ConcurrentHashMap<String, RefundPaymentResponse>()

    // Simulate transient failures: every 5th call fails with UNAVAILABLE
    private var callCount = 0

    override suspend fun chargePayment(request: ChargePaymentRequest): ChargePaymentResponse {
        // Validation — INVALID_ARGUMENT is never retryable
        if (request.amountCents <= 0) {
            throw StatusException(
                Status.INVALID_ARGUMENT.withDescription("amount_cents must be positive")
            )
        }
        if (request.currency.isBlank()) {
            throw StatusException(
                Status.INVALID_ARGUMENT.withDescription("currency is required")
            )
        }

        // Idempotency check — return existing result if key already seen
        idempotencyMap[request.idempotencyKey]?.let { existing ->
            log.info { "Idempotent charge: returning existing txId=${existing.transactionId}" }
            return existing
        }

        // Simulate transient failure (demonstrates circuit breaker behavior)
        callCount++
        if (callCount % 5 == 0) {
            throw StatusException(
                Status.UNAVAILABLE.withDescription("Payment gateway temporarily unavailable")
            )
        }

        val txId = UUID.randomUUID().toString()
        val response = chargePaymentResponse {
            transactionId = txId
            chargedAt     = Instant.now().toProtoTimestamp()
        }

        transactions[txId] = response
        idempotencyMap[request.idempotencyKey] = response

        log.info { "Payment charged: orderId=${request.orderId} txId=$txId amount=${request.amountCents}${request.currency}" }
        return response
    }

    override suspend fun refundPayment(request: RefundPaymentRequest): RefundPaymentResponse {
        transactions[request.transactionId]
            ?: throw StatusException(
                Status.NOT_FOUND.withDescription("Transaction not found: ${request.transactionId}")
            )

        val refundId = UUID.randomUUID().toString()
        val response = refundPaymentResponse {
            this.refundId  = refundId
            refundedAt     = Instant.now().toProtoTimestamp()
        }
        refunds[refundId] = response
        log.info { "Refunded txId=${request.transactionId} → refundId=$refundId reason=${request.reason}" }
        return response
    }
}

fun main() {
    val port = System.getenv("GRPC_PORT")?.toInt() ?: 50052
    val server = ServerBuilder.forPort(port)
        .addService(PaymentGrpcService())
        .build()
        .start()
    log.info { "Payment Service started on port $port" }
    server.awaitTermination()
}

private fun Instant.toProtoTimestamp(): Timestamp =
    Timestamp.newBuilder().setSeconds(epochSecond).setNanos(nano).build()
