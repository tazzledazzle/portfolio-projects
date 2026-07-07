package com.skidroad.saga.orchestrator.resilience

import io.grpc.Status
import io.grpc.StatusException
import kotlinx.coroutines.delay
import mu.KotlinLogging
import java.util.concurrent.atomic.AtomicInteger
import java.util.concurrent.atomic.AtomicLong
import java.util.concurrent.atomic.AtomicReference

private val log = KotlinLogging.logger {}

/**
 * Coroutine-friendly Circuit Breaker implementing the classic three-state FSM:
 *
 *   CLOSED ──(failure threshold exceeded)──► OPEN
 *     ▲                                        │
 *     └────(probe succeeds)──── HALF_OPEN ◄───┘
 *                                        │
 *              (probe fails)─────────────┘ (back to OPEN)
 *
 * Key design decisions:
 *  - Uses AtomicReference for state — no synchronized blocks, safe for coroutines
 *  - Wraps gRPC [StatusException] for seamless integration with gRPC stubs
 *  - [OPEN] state surfaces as [Status.UNAVAILABLE] so callers can apply retry logic
 *
 * @param name            Service name for logging (e.g. "payment-service")
 * @param failureThreshold Number of consecutive failures before opening
 * @param resetTimeoutMs  How long to stay OPEN before probing (half-open)
 */
class CircuitBreaker(
    val name: String,
    private val failureThreshold: Int = 5,
    private val resetTimeoutMs:   Long = 10_000L
) {
    enum class State { CLOSED, OPEN, HALF_OPEN }

    private val state           = AtomicReference(State.CLOSED)
    private val failureCount    = AtomicInteger(0)
    private val lastFailureTime = AtomicLong(0L)

    val currentState: State get() = state.get()

    /**
     * Execute [block] through the circuit breaker.
     * Throws [StatusException] with [Status.UNAVAILABLE] if the circuit is OPEN.
     */
    suspend fun <T> execute(block: suspend () -> T): T {
        when (state.get()) {
            State.OPEN -> {
                val elapsed = System.currentTimeMillis() - lastFailureTime.get()
                if (elapsed >= resetTimeoutMs) {
                    log.info { "[$name] Circuit breaker: OPEN → HALF_OPEN (probing)" }
                    state.set(State.HALF_OPEN)
                } else {
                    throw StatusException(
                        Status.UNAVAILABLE.withDescription(
                            "Circuit breaker OPEN for $name — retry after ${resetTimeoutMs - elapsed}ms"
                        )
                    )
                }
            }
            else -> { /* CLOSED or HALF_OPEN: proceed */ }
        }

        return try {
            val result = block()
            onSuccess()
            result
        } catch (e: StatusException) {
            onFailure(e)
            throw e
        }
    }

    private fun onSuccess() {
        if (state.get() == State.HALF_OPEN) {
            log.info { "[$name] Circuit breaker: HALF_OPEN → CLOSED (probe succeeded)" }
        }
        state.set(State.CLOSED)
        failureCount.set(0)
    }

    private fun onFailure(e: StatusException) {
        lastFailureTime.set(System.currentTimeMillis())
        val count = failureCount.incrementAndGet()
        log.warn { "[$name] Circuit breaker failure $count/$failureThreshold: ${e.status}" }

        if (state.get() == State.HALF_OPEN || count >= failureThreshold) {
            log.error { "[$name] Circuit breaker: → OPEN" }
            state.set(State.OPEN)
        }
    }
}

/**
 * Retry policy with exponential backoff and jitter.
 * Only retries on [Status.UNAVAILABLE] and [Status.DEADLINE_EXCEEDED] —
 * never on [Status.INVALID_ARGUMENT], [Status.NOT_FOUND], etc. (non-retryable).
 *
 * @param maxAttempts   Total attempts (1 = no retry)
 * @param baseDelayMs   Initial delay before first retry
 * @param maxDelayMs    Cap on delay (prevents unbounded growth)
 * @param jitterFactor  Adds randomness: delay *= (1 ± jitterFactor)
 */
class RetryPolicy(
    private val maxAttempts:  Int  = 3,
    private val baseDelayMs:  Long = 100L,
    private val maxDelayMs:   Long = 5_000L,
    private val jitterFactor: Double = 0.2
) {
    private val retryableStatuses = setOf(
        Status.Code.UNAVAILABLE,
        Status.Code.DEADLINE_EXCEEDED,
        Status.Code.RESOURCE_EXHAUSTED
    )

    suspend fun <T> execute(operationName: String, block: suspend () -> T): T {
        var lastException: StatusException? = null

        repeat(maxAttempts) { attempt ->
            try {
                return block()
            } catch (e: StatusException) {
                lastException = e
                if (e.status.code !in retryableStatuses) {
                    log.debug { "[$operationName] Non-retryable status ${e.status.code}, failing fast" }
                    throw e
                }
                if (attempt < maxAttempts - 1) {
                    val backoff = (baseDelayMs * Math.pow(2.0, attempt.toDouble())).toLong()
                        .coerceAtMost(maxDelayMs)
                    val jitter  = (backoff * jitterFactor * (Math.random() * 2 - 1)).toLong()
                    val delay   = (backoff + jitter).coerceAtLeast(0)
                    log.warn { "[$operationName] Attempt ${attempt + 1}/$maxAttempts failed (${e.status.code}), retrying in ${delay}ms" }
                    delay(delay)
                }
            }
        }

        throw lastException!!
    }
}
