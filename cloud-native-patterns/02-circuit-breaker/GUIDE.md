# Circuit Breakers in Practice: Resilience4j and Istio DestinationRule

A circuit breaker is a proxy that tracks the failure rate of calls to a downstream dependency and stops sending traffic when that dependency is clearly failing. It's named after the electrical device: when a circuit is overloaded, the breaker trips and cuts the current before the wiring burns out. In software, the "wiring" is your threads, memory, and connection pools — and the "burn" is a cascading failure that takes down your entire service along with the one that was already broken.

This guide walks through a two-layer implementation: application-level circuit breaking with Resilience4j in Kotlin, and mesh-level outlier detection with an Istio DestinationRule.

---

## Why Cascading Failures Happen

Imagine `orders-service` calls `payment-service`, which is responding slowly — say, 10-second timeouts on every request. If `orders-service` has a thread pool of 50 threads and each request waits 10 seconds, 50 in-flight requests exhaust the pool in seconds. New incoming requests to `orders-service` start queuing, then timing out. The problem has propagated upstream.

The circuit breaker pattern breaks this chain. Once enough calls to `payment-service` have failed, the breaker opens and subsequent calls fail immediately — in microseconds — instead of waiting for a timeout. This frees the thread pool to handle requests that don't depend on `payment-service`, limits error propagation, and gives `payment-service` time to recover without being hammered by retries.

---

## The Three States

A circuit breaker cycles through three states:

**CLOSED** — Normal operation. Calls pass through. The breaker records successes and failures in a sliding window. When the failure rate exceeds the threshold, it transitions to OPEN.

**OPEN** — Fast-fail mode. All calls are rejected immediately with a `CallNotPermittedException` (no network call made). After a configured wait duration, the breaker transitions to HALF_OPEN.

**HALF_OPEN** — Probe mode. A limited number of calls are allowed through to test whether the downstream has recovered. If enough succeed, the breaker closes. If too many fail, it opens again.

---

## Application-Level Implementation: Resilience4j

The `buildCircuitBreakerConfig()` function defines the tuning parameters:

```kotlin
fun buildCircuitBreakerConfig(): CircuitBreakerConfig =
    CircuitBreakerConfig.custom()
        .failureRateThreshold(50f)
        .waitDurationInOpenState(Duration.ofSeconds(30))
        .slidingWindowType(CircuitBreakerConfig.SlidingWindowType.COUNT_BASED)
        .slidingWindowSize(20)
        .permittedNumberOfCallsInHalfOpenState(5)
        .automaticTransitionFromOpenToHalfOpenEnabled(true)
        .recordExceptions(TransientPaymentException::class.java)
        .ignoreExceptions(PaymentDeclinedException::class.java)
        .build()
```

Breaking down the key parameters:

- **`failureRateThreshold(50f)`**: The breaker opens when 50% or more of the last 20 calls (the sliding window) were failures.
- **`slidingWindowType(COUNT_BASED)`**: Uses the last N calls, not a time window. COUNT_BASED is more predictable under variable load; TIME_BASED is better when call rate varies significantly across time windows.
- **`slidingWindowSize(20)`**: 20-call window. Too small and the breaker is trigger-happy; too large and it reacts slowly to sustained failure.
- **`waitDurationInOpenState(30s)`**: How long the breaker stays OPEN before probing with HALF_OPEN. Shorter means faster recovery but also means hitting a recovering service more aggressively.
- **`permittedNumberOfCallsInHalfOpenState(5)`**: 5 probe calls. If fewer than 50% fail, the breaker closes.
- **`automaticTransitionFromOpenToHalfOpenEnabled(true)`**: Without this, the breaker stays OPEN forever until a call is made. With it, the transition happens automatically on a background thread after the wait duration.

### Exception Classification

This is the most important configuration detail:

```kotlin
.recordExceptions(TransientPaymentException::class.java)
.ignoreExceptions(PaymentDeclinedException::class.java)
```

`TransientPaymentException` represents a network or infrastructure failure — the payment gateway is unreachable, connection reset, 5xx from the payment service. These count toward the failure rate.

`PaymentDeclinedException` is a business error — the customer's card was declined. This is not a transient failure. The payment service is working correctly; it just said no. Counting this as a circuit-breaker failure would open the breaker when payments are being legitimately declined, which would take down legitimate charges along with them.

The distinction between infrastructure failures and business errors is critical for any circuit-breaker configuration. Get it wrong and you'll trip the breaker during peak sales when declined cards are more common.

### Decoration Order Matters

```kotlin
fun chargeCard(request: ChargeRequest): ChargeResult {
    return Decorators.ofSupplier { client.charge(request) }
        .withCircuitBreaker(circuitBreaker)
        .withRetry(retry)
        .decorate()
        .get()
}
```

The decoration order is `retry(circuitBreaker(call))`. This means the retry wraps the circuit-breaker-protected call. If the circuit breaker is OPEN, the call fails immediately with `CallNotPermittedException`, and the retry does **not** retry it — which is correct. The point of an open circuit breaker is to stop sending traffic. Retrying OPEN rejections would defeat the purpose.

The inverse order — `circuitBreaker(retry(call))` — would let the retry exhaust its attempts before the circuit breaker even sees the result, defeating the fast-fail behavior.

### Observability

```kotlin
circuitBreaker.eventPublisher
    .onStateTransition { event ->
        log.info("[CircuitBreaker] State transition: {} → {}",
            event.stateTransition.fromState,
            event.stateTransition.toState)
    }
    .onFailureRateExceeded { event ->
        log.warn("[CircuitBreaker] Failure rate threshold exceeded: {:.1f}%",
            event.failureRate)
    }
    .onCallNotPermitted { _ ->
        log.warn("[CircuitBreaker] Call rejected — circuit is OPEN")
    }
```

State transitions are the signal you alert on. In production, these log lines should also emit metrics (Prometheus gauge for state, counter for rejections). A circuit breaker that trips silently is worse than no circuit breaker — you won't know the downstream is failing until upstream services start showing symptoms.

### Retry with Exponential Backoff

```kotlin
fun buildRetryConfig(): RetryConfig =
    RetryConfig.custom<ChargeResult>()
        .maxAttempts(3)
        .intervalFunction(
            IntervalFunction.ofExponentialRandomBackoff(
                Duration.ofMillis(100),
                2.0,
                0.5,
            )
        )
        .retryExceptions(TransientPaymentException::class.java)
        .build()
```

The `ofExponentialRandomBackoff` function adds ±50% jitter to each interval. Base 100ms, multiplier 2.0 produces ~100ms, ~200ms, ~400ms intervals — but with jitter, so not all clients back off simultaneously. Without jitter, a wave of clients all hitting a recovering service at exactly the same intervals creates a thundering-herd problem.

---

## Mesh-Level Circuit Breaking: Istio DestinationRule

The Resilience4j configuration handles application-level failures — scenarios where the payment service returns a 200 OK with an error payload, or throws a business exception. But it can't handle what Istio can: pure network-level failure signals.

The DestinationRule for `payment-service` adds a second layer:

```yaml
trafficPolicy:
  connectionPool:
    tcp:
      maxConnections: 100
      connectTimeout: 3s
    http:
      http1MaxPendingRequests: 50
      http2MaxRequests: 1000
      maxRequestsPerConnection: 10
      maxRetries: 3
  outlierDetection:
    consecutive5xxErrors: 5
    consecutiveGatewayErrors: 5
    interval: 30s
    baseEjectionTime: 30s
    maxEjectionPercent: 50
    minHealthPercent: 30
```

### Connection Pool Limits

`http1MaxPendingRequests: 50` is the first line of defense. If `payment-service` pods are all slow, requests queued beyond 50 get a 503 immediately rather than queuing indefinitely. This is what actually prevents thread pool exhaustion in `orders-service`. Without this limit, a slow downstream can absorb all your threads across hundreds of connections.

`maxRequestsPerConnection: 10` prevents connection monopolization — one slow request won't hold a connection open and starve others.

### Outlier Detection

```yaml
outlierDetection:
  consecutive5xxErrors: 5
  baseEjectionTime: 30s
  maxEjectionPercent: 50
  minHealthPercent: 30
```

Outlier detection ejects specific pods from the load-balancing pool when they return too many errors. After 5 consecutive 5xx responses from a particular `payment-service` pod, that pod is ejected for at least 30 seconds. It won't receive traffic until the ejection time expires and Istio probes it.

`maxEjectionPercent: 50` is a safety valve — Istio won't eject more than half the pod pool at once. Without this, a brief thunderstorm of errors could trigger ejection of all pods, taking down the service entirely.

`minHealthPercent: 30` is the floor: if fewer than 30% of pods are healthy, Istio stops ejecting regardless of error rates, preferring degraded service over complete outage.

---

## App Layer vs. Mesh Layer: What Each Handles

| Concern | Resilience4j (App) | Istio DestinationRule (Mesh) |
|---------|-------------------|------------------------------|
| Business error classification | Yes — ignore PaymentDeclined | No — only sees HTTP status |
| Individual pod ejection | No — routes to any pod | Yes — per-pod outlier detection |
| Connection pool limits | No | Yes |
| Retry with backoff + jitter | Yes | Yes (simpler, no jitter) |
| State visibility in logs/metrics | Yes | Yes (via Envoy access logs) |
| Cross-language | No — per service | Yes — all services in mesh |

The two layers are complementary, not redundant. The app layer makes semantic distinctions (business vs. infra error) that the mesh can't. The mesh layer enforces connection budgets and removes misbehaving pods from rotation — something the app layer can't do because it doesn't control routing.

The main risk of layering: **retry amplification**. If Resilience4j retries 3 times and Istio's `maxRetries: 3` also retries 3 times, a single failed call can generate 9 upstream requests. In this codebase, Istio retries are set conservatively, and the `retryOn` condition (`5xx,reset,connect-failure,retriable-4xx`) is specific. Review these together when tuning.

---

## Key Takeaways

- Circuit breakers protect upstream services from cascading failure caused by slow or failing downstreams — fast-fail is better than slow-fail
- Three states: CLOSED (normal), OPEN (fast-fail), HALF_OPEN (probing) — the transition from OPEN to HALF_OPEN should be automatic in production
- Classify exceptions carefully: infrastructure failures count toward the failure rate; business errors should be ignored
- Decoration order `retry(circuitBreaker(call))` is correct — retrying OPEN rejections defeats the pattern
- Add ±jitter to retry intervals to avoid synchronized thundering herds
- Istio DestinationRule adds a second layer at the mesh level: connection pool limits (prevent thread exhaustion) and outlier detection (eject bad pods)
- Beware retry amplification when combining app-level and mesh-level retry — total retries multiply
- Observable state transitions are as important as the circuit breaker itself — alert on state changes, not just on errors
- `maxEjectionPercent` and `minHealthPercent` prevent outlier detection from taking down an entire service during a transient error storm
