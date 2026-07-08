package com.patterns.circuitbreaker

import org.slf4j.LoggerFactory
import java.io.IOException
import java.util.concurrent.atomic.AtomicInteger

data class ChargeRequest(
    val orderId: String,
    val amountCents: Long,
    val idempotencyKey: String,
)

data class ChargeResult(
    val transactionId: String,
    val status: ChargeStatus,
)

enum class ChargeStatus { SUCCESS, DECLINED }

/**
 * Stub payment client that simulates a flaky downstream dependency.
 *
 * Failure behavior is controlled by [failEveryN]: every Nth call throws an
 * [IOException] to simulate a network-level failure. This is the class of error
 * the circuit breaker and retry logic are designed to handle — transient failures
 * that are worth retrying, as opposed to business errors (DECLINED) which should
 * not be retried.
 */
class PaymentClient(private val failEveryN: Int = 3) {

    private val log = LoggerFactory.getLogger(PaymentClient::class.java)
    private val callCount = AtomicInteger(0)

    /**
     * Simulates a charge call. Throws [IOException] every [failEveryN] calls.
     *
     * Note: in production, this would be an HTTP/gRPC call to the payment service.
     * The IOException simulates connection reset, timeout, or 5xx.
     */
    fun charge(request: ChargeRequest): ChargeResult {
        val n = callCount.incrementAndGet()
        log.debug("PaymentClient.charge() call #{} orderId={}", n, request.orderId)

        if (n % failEveryN == 0) {
            log.warn("Simulating failure on call #{}", n)
            throw IOException("Simulated network failure on call #$n (failEveryN=$failEveryN)")
        }

        return ChargeResult(
            transactionId = "txn-${request.orderId}-$n",
            status = ChargeStatus.SUCCESS,
        )
    }

    fun resetCallCount() {
        callCount.set(0)
    }

    fun getCallCount(): Int = callCount.get()
}
