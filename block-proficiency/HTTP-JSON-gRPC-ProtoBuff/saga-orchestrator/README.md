# Saga Orchestrator — Kotlin gRPC Demo

Demonstrates: gRPC interceptors (auth, tracing, logging), deadline propagation,
circuit breaking, retry with jitter, Protobuf well-known types, service reflection,
and the Saga compensation pattern.

## Architecture

```
Client
  │
  ▼ (gRPC / REST via gRPC-Gateway)
Orchestrator :50051
  │  Interceptor chain: Auth → Tracing → Logging
  │  CircuitBreaker + RetryPolicy per downstream
  ├──► Payment Service  :50052
  ├──► Inventory Service :50053
  └──► Shipping Service  :50054  (30% fail rate by default)
```

## Quick Start

```bash
# Build all modules
./gradlew build

# Start all services
docker compose up

# Or run each service in separate terminals:
./gradlew :payment-service:run
./gradlew :inventory-service:run
./gradlew :shipping-service:run
./gradlew :orchestrator:run
```

## grpcurl Commands

### 1. Discover services via server reflection
```bash
grpcurl -plaintext \
  -H 'Authorization: Bearer dev-token' \
  localhost:50051 list
# → saga.v1.SagaOrchestratorService
# → grpc.reflection.v1alpha.ServerReflection

grpcurl -plaintext \
  -H 'Authorization: Bearer dev-token' \
  localhost:50051 describe saga.v1.SagaOrchestratorService
```

### 2. Place an order (happy path — run a few times, shipping will eventually fail)
```bash
grpcurl -plaintext \
  -H 'Authorization: Bearer dev-token' \
  -d '{
    "order_id": "ord-001",
    "idempotency_key": "idem-001",
    "items": [{"sku": "SKU-001", "quantity": 2, "price_cents": 999}],
    "payment_info": {
      "payment_method_id": "pm_test_visa",
      "amount_cents": 1998,
      "currency": "USD"
    },
    "address": {
      "street": "123 Yesler Way",
      "city": "Seattle",
      "postcode": "98104",
      "country": "US"
    }
  }' \
  localhost:50051 saga.v1.SagaOrchestratorService/PlaceOrder
```

### 3. Check saga status
```bash
grpcurl -plaintext \
  -H 'Authorization: Bearer dev-token' \
  -d '{"saga_id": "<sagaId from response above>"}' \
  localhost:50051 saga.v1.SagaOrchestratorService/GetSagaStatus
```

### 4. Trigger auth failure (no token)
```bash
grpcurl -plaintext \
  -d '{"order_id": "ord-002"}' \
  localhost:50051 saga.v1.SagaOrchestratorService/PlaceOrder
# → ERROR: Code=Unauthenticated Desc=Bearer token required
```

### 5. Force shipping failure to observe compensation
```bash
# Set SHIPPING_FAIL_RATE=1.0 in docker-compose.yml and restart shipping-service
# Then place an order — you'll see compensation log in the error response:
# "Compensation: inventory=OK, payment=OK"
```

### 6. Observe circuit breaker opening
```bash
# Run PlaceOrder ~5 times with SHIPPING_FAIL_RATE=1.0
# After 5 failures, subsequent calls fail immediately with:
# → Code=Unavailable Desc=Circuit breaker OPEN for shipping-service
```

## Key Design Decisions

### Interceptor ordering
Interceptors are registered Auth → Tracing → Logging on ServerBuilder.
gRPC applies them last-registered-first on inbound requests, so the actual
execution order is Auth (outermost) → Tracing → Logging → handler.
This ensures trace-id and principal are in Context by the time LoggingInterceptor reads them.

### Deadline propagation
Each downstream stub call uses `.withDeadlineAfter(N, SECONDS)`. This sets a
per-call deadline, not a global one. The values (payment=3s, inventory=3s, shipping=4s)
sum to 10s — matching the conceptual 10s parent deadline from the client.
In practice, use Context.current().deadline to propagate the *remaining* parent
deadline rather than a fixed value.

### Idempotency
Every request includes an idempotency_key. Downstream services store
key → response and return the cached response on replay. This makes the entire
saga safe to retry at any step without double-charging.

### Compensation ordering
Compensation runs in reverse step order: shipping (not needed — it failed) →
inventory → payment. This is critical: releasing inventory before refunding payment
prevents a race where inventory is re-reserved by another order against a
payment that hasn't been refunded yet.

### Protobuf well-known types
- `Timestamp`: used for charged_at, refunded_at, estimated_delivery
- `Duration`: used for reservation_ttl (15 minutes) — semantically cleaner than int64 epoch
- `Any`: available in CreateShipmentResponse for carrier-specific metadata (see proto comment)

## Running Tests

```bash
./gradlew :orchestrator:test
```

Tests use `InProcessServerBuilder` — full gRPC serialization, no network.
Covers: happy path, compensation, circuit breaker FSM, retry policy, idempotency.
