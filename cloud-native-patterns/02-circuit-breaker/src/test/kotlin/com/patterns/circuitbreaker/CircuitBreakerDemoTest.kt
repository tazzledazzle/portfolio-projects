package com.patterns.circuitbreaker

import io.github.resilience4j.circuitbreaker.CircuitBreaker
import io.github.resilience4j.circuitbreaker.CircuitBreakerConfig
import io.github.resilience4j.retry.Retry
import io.github.resilience4j.retry.RetryConfig
import io.mockk.every
import io.mockk.mockk
import org.assertj.core.api.Assertions.assertThat
import org.assertj.core.api.Assertions.assertThatThrownBy
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import java.time.Duration
import java.util.UUID
import java.util.concurrent.atomic.AtomicInteger

class CircuitBreakerDemoTest {

    private lateinit var mockClient: PaymentClient
    private lateinit var service: PaymentService

    // Tight config for fast test execution — no 30s wait times.
    private val testCbConfig: CircuitBreakerConfig = CircuitBreakerConfig.custom()
        .failureRateThreshold(50f)
        .waitDurationInOpenState(Duration.ofMillis(200))
        .slidingWindowType(CircuitBreakerConfig.SlidingWindowType.COUNT_BASED)
        .slidingWindowSize(4)
        .permittedNumberOfCallsInHalfOpenState(2)
        .automaticTransitionFromOpenToHalfOpenEnabled(true)
        .recordExceptions(TransientPaymentException::class.java)
        .build()

    private val testRetryConfig: RetryConfig = RetryConfig.custom<ChargeResult>()
        .maxAttempts(3)
        .waitDuration(Duration.ofMillis(10))
        .retryExceptions(TransientPaymentException::class.java)
        .build()

    private val successResult = ChargeResult(
        transactionId = "txn-test-success",
        status = ChargeStatus.SUCCESS,
    )

    private fun request(orderId: String = "order-test") = ChargeRequest(
        orderId = orderId,
        amountCents = 100L,
        idempotencyKey = UUID.randomUUID().toString(),
    )

    @BeforeEach
    fun setUp() {
        mockClient = mockk<PaymentClient>()
        service = PaymentService(mockClient, testCbConfig, testRetryConfig)
    }

    @Test
    fun `circuit breaker starts CLOSED`() {
        assertThat(service.getState()).isEqualTo(CircuitBreaker.State.CLOSED)
    }

    @Test
    fun `successful calls keep breaker CLOSED`() {
        every { mockClient.charge(any()) } returns successResult

        repeat(6) { service.chargeCard(request("order-$it")) }

        assertThat(service.getState()).isEqualTo(CircuitBreaker.State.CLOSED)
    }

    @Test
    fun `breaker opens after failure rate exceeds threshold`() {
        // With slidingWindowSize=4 and threshold=50%, breaker opens after 2/4 failures.
        // Retry is configured with maxAttempts=2, so each "call" below produces 2 actual
        // attempts to the mock before propagating the error. That means 2 failures in the
        // sliding window come from 1 failed chargeCard() invocation (both attempts recorded).
        every { mockClient.charge(any()) } throws TransientPaymentException("downstream unavailable")

        // Collect state transitions
        val transitions = mutableListOf<Pair<CircuitBreaker.State, CircuitBreaker.State>>()
        service.circuitBreaker.eventPublisher.onStateTransition { event ->
            transitions.add(event.stateTransition.fromState to event.stateTransition.toState)
        }

        // Drive enough failures to fill the sliding window and exceed the threshold
        repeat(4) {
            runCatching { service.chargeCard(request("order-fail-$it")) }
        }

        assertThat(service.getState()).isEqualTo(CircuitBreaker.State.OPEN)
        assertThat(transitions).anyMatch { (from, to) ->
            from == CircuitBreaker.State.CLOSED && to == CircuitBreaker.State.OPEN
        }
    }

    @Test
    fun `open breaker rejects calls immediately without calling downstream`() {
        every { mockClient.charge(any()) } throws TransientPaymentException("downstream unavailable")

        // Open the breaker
        repeat(4) { runCatching { service.chargeCard(request("order-open-$it")) } }
        assertThat(service.getState()).isEqualTo(CircuitBreaker.State.OPEN)

        // Now the client should NOT be called — the breaker short-circuits
        every { mockClient.charge(any()) } throws AssertionError("should not be called when OPEN")

        assertThatThrownBy { service.chargeCard(request("order-rejected")) }
            .isInstanceOf(io.github.resilience4j.circuitbreaker.CallNotPermittedException::class.java)
    }

    @Test
    fun `breaker transitions OPEN → HALF_OPEN → CLOSED on recovery`() {
        val transitions = mutableListOf<Pair<CircuitBreaker.State, CircuitBreaker.State>>()
        service.circuitBreaker.eventPublisher.onStateTransition { event ->
            transitions.add(event.stateTransition.fromState to event.stateTransition.toState)
        }

        // Phase 1: fail until OPEN
        every { mockClient.charge(any()) } throws TransientPaymentException("timeout")
        repeat(4) { runCatching { service.chargeCard(request("order-fail-$it")) } }
        assertThat(service.getState()).isEqualTo(CircuitBreaker.State.OPEN)

        // Phase 2: wait for auto-transition to HALF_OPEN (waitDuration=200ms in test config)
        Thread.sleep(300)
        // Trigger a call to force the auto-transition evaluation
        every { mockClient.charge(any()) } returns successResult
        runCatching { service.chargeCard(request("half-open-probe-1")) }
        runCatching { service.chargeCard(request("half-open-probe-2")) }

        // Phase 3: with permittedNumberOfCallsInHalfOpenState=2 and both succeeding, breaker closes
        assertThat(service.getState()).isEqualTo(CircuitBreaker.State.CLOSED)

        // Verify the full state transition sequence
        val fromStates = transitions.map { it.first }
        assertThat(fromStates).contains(CircuitBreaker.State.CLOSED)   // CLOSED → OPEN
        assertThat(fromStates).contains(CircuitBreaker.State.OPEN)     // OPEN → HALF_OPEN
    }

    @Test
    fun `retry fires on transient failure before counting against circuit breaker`() {
        // Test retry in isolation — no MockK, no circuit breaker.
        // Confirms that the Resilience4j Retry decorator calls the supplier again on IOException.
        val retryOnlyConfig = RetryConfig.custom<ChargeResult>()
            .maxAttempts(3)
            .waitDuration(Duration.ofMillis(10))
            .retryExceptions(TransientPaymentException::class.java)
            .build()
        val retryOnly = Retry.of("isolated-retry", retryOnlyConfig)

        val callCount = AtomicInteger(0)
        val decorated = Retry.decorateSupplier(retryOnly) {
            val n = callCount.incrementAndGet()
            if (n == 1) throw TransientPaymentException("first attempt fails")
            successResult
        }

        val result = decorated.get()

        assertThat(result.status).isEqualTo(ChargeStatus.SUCCESS)
        assertThat(callCount.get()).isEqualTo(2)
        // Circuit breaker state is irrelevant — this test covers retry layer only
    }

    @Test
    fun `business exceptions are not retried and do not open the circuit breaker`() {
        every { mockClient.charge(any()) } throws PaymentDeclinedException("Card declined")

        // Fill the sliding window with declined payments
        repeat(4) {
            assertThatThrownBy { service.chargeCard(request("order-declined-$it")) }
                .isInstanceOf(PaymentDeclinedException::class.java)
        }

        // Breaker must remain CLOSED — declines are not failures from the breaker's perspective
        assertThat(service.getState()).isEqualTo(CircuitBreaker.State.CLOSED)
    }
}
