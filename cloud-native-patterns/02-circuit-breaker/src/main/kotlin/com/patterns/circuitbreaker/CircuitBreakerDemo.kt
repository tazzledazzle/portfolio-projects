package com.patterns.circuitbreaker

import io.github.resilience4j.circuitbreaker.CircuitBreaker
import io.github.resilience4j.circuitbreaker.CircuitBreakerConfig
import io.github.resilience4j.circuitbreaker.CircuitBreakerRegistry
import io.github.resilience4j.core.IntervalFunction
import io.github.resilience4j.decorators.Decorators
import io.github.resilience4j.retry.Retry
import io.github.resilience4j.retry.RetryConfig
import io.github.resilience4j.retry.RetryRegistry
import org.slf4j.LoggerFactory
import java.io.IOException
import java.time.Duration
import java.util.UUID

private val log = LoggerFactory.getLogger("CircuitBreakerDemo")

/**
 * Builds the circuit breaker configuration.
 *
 * Key parameters:
 *  - 50% failure rate threshold — breaker opens when ≥50% of last 20 calls fail
 *  - COUNT_BASED sliding window of 20 — uses the last 20 calls (not a time window)
 *  - 30s wait in OPEN state before transitioning to HALF_OPEN
 *  - 5 permitted calls in HALF_OPEN — breaker closes if <50% of these fail
 */
fun buildCircuitBreakerConfig(): CircuitBreakerConfig =
    CircuitBreakerConfig.custom()
        .failureRateThreshold(50f)
        .waitDurationInOpenState(Duration.ofSeconds(30))
        .slidingWindowType(CircuitBreakerConfig.SlidingWindowType.COUNT_BASED)
        .slidingWindowSize(20)
        .permittedNumberOfCallsInHalfOpenState(5)
        .automaticTransitionFromOpenToHalfOpenEnabled(true)
        // Only count IOExceptions as failures — not business-logic errors like
        // payment declined. This is a critical distinction: don't circuit-break
        // on 4xx-equivalent business errors.
        .recordExceptions(IOException::class.java)
        .ignoreExceptions(PaymentDeclinedException::class.java)
        .build()

/**
 * Builds the retry configuration with exponential backoff + jitter.
 *
 * Jitter prevents synchronized retry storms when many clients back off together.
 * ofExponentialRandomBackoff adds ±50% random jitter to each interval by default.
 */
fun buildRetryConfig(): RetryConfig =
    RetryConfig.custom<ChargeResult>()
        .maxAttempts(3)
        .intervalFunction(
            // Base 100ms, multiplier 2.0, randomization factor 0.5
            // Resulting intervals: ~100ms, ~200ms, ~400ms (with ±50% jitter)
            IntervalFunction.ofExponentialRandomBackoff(
                Duration.ofMillis(100),
                2.0,
                0.5,
            )
        )
        // Only retry transient I/O errors — never retry a declined charge.
        .retryOnException { it is IOException }
        .build()

/**
 * Business exception — should NOT be retried or counted as a circuit-breaker failure.
 * This is the kind of error the mesh-level breaker can't distinguish from a real failure.
 */
class PaymentDeclinedException(message: String) : RuntimeException(message)

/**
 * Service that wraps [PaymentClient] with circuit breaker + retry.
 *
 * The decoration order matters:
 *   retry(circuitBreaker(call))
 * This means: the retry wraps the circuit-breaker-protected call. If the circuit
 * breaker is OPEN, the call fails immediately (throws CallNotPermittedException)
 * and the retry does NOT retry it — which is correct, since OPEN means the
 * downstream is known-bad and we shouldn't hammer it.
 */
class PaymentService(
    private val client: PaymentClient,
    circuitBreakerConfig: CircuitBreakerConfig = buildCircuitBreakerConfig(),
    retryConfig: RetryConfig = buildRetryConfig(),
) {
    val circuitBreaker: CircuitBreaker
    private val retry: Retry

    init {
        val cbRegistry = CircuitBreakerRegistry.of(circuitBreakerConfig)
        circuitBreaker = cbRegistry.circuitBreaker("payment-service")

        val retryRegistry = RetryRegistry.of(retryConfig)
        retry = retryRegistry.retry("payment-service")

        // Log all state transitions — in production these feed into metrics/alerts.
        circuitBreaker.eventPublisher
            .onStateTransition { event ->
                log.info(
                    "[CircuitBreaker] State transition: {} → {}",
                    event.stateTransition.fromState,
                    event.stateTransition.toState,
                )
            }
            .onFailureRateExceeded { event ->
                log.warn(
                    "[CircuitBreaker] Failure rate threshold exceeded: {:.1f}%",
                    event.failureRate,
                )
            }
            .onCallNotPermitted { _ ->
                log.warn("[CircuitBreaker] Call rejected — circuit is OPEN")
            }

        retry.eventPublisher
            .onRetry { event ->
                log.info(
                    "[Retry] Attempt #{} for '{}' — last error: {}",
                    event.numberOfRetryAttempts,
                    event.name,
                    event.lastThrowable?.message,
                )
            }
    }

    fun chargeCard(request: ChargeRequest): ChargeResult {
        return Decorators.ofSupplier { client.charge(request) }
            .withCircuitBreaker(circuitBreaker)
            .withRetry(retry)
            .decorate()
            .get()
    }

    fun getState(): CircuitBreaker.State = circuitBreaker.state
}

fun main() {
    log.info("=== Circuit Breaker + Retry Demo ===")
    log.info("")
    log.info("Config: 50% failure threshold, COUNT_BASED window of 20, failEveryN=3 in client")
    log.info("Expected: circuit opens after ~10 failures in 20 calls")
    log.info("")

    // failEveryN=3 means ~33% of calls fail — well below 50% threshold.
    // We'll use failEveryN=2 (50%) to reliably trigger the breaker.
    val client = PaymentClient(failEveryN = 2)
    val service = PaymentService(client)

    var successCount = 0
    var failureCount = 0
    var openRejections = 0

    // Run 40 calls. With 50% failure rate, the breaker should open around call 20.
    repeat(40) { i ->
        val orderId = "order-${i + 1}"
        val request = ChargeRequest(
            orderId = orderId,
            amountCents = 9999L,
            idempotencyKey = UUID.randomUUID().toString(),
        )

        try {
            val result = service.chargeCard(request)
            successCount++
            log.info("[Call ${i + 1}] SUCCESS txnId=${result.transactionId} state=${service.getState()}")
        } catch (e: io.github.resilience4j.circuitbreaker.CallNotPermittedException) {
            openRejections++
            log.warn("[Call ${i + 1}] OPEN — circuit breaker rejected call (fast-fail)")
        } catch (e: Exception) {
            failureCount++
            log.error("[Call ${i + 1}] FAILED after retries: {} state={}", e.message, service.getState())
        }

        // Brief pause so logs are readable
        Thread.sleep(50)
    }

    log.info("")
    log.info("=== Results ===")
    log.info("Total calls attempted : 40")
    log.info("Successes             : $successCount")
    log.info("Failures (after retry): $failureCount")
    log.info("Open rejections       : $openRejections")
    log.info("Final CB state        : ${service.getState()}")
    log.info("")
    log.info("State transitions logged above show CLOSED → OPEN.")
    log.info("After 30s wait, CB auto-transitions to HALF_OPEN for probe calls.")
}
