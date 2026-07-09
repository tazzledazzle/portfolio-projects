# Distributed Transactions with the Saga Pattern and Temporal

Placing an order involves multiple services: charge the payment, reserve the inventory, schedule the shipment. Each of these is a separate service with its own database. If the shipment scheduling fails after the payment has been charged and inventory reserved, you need to undo those previous steps. This is a distributed transaction — a sequence of operations across multiple services that must either all succeed or all be rolled back.

The classical solution is Two-Phase Commit (2PC). The modern solution for microservices is the **Saga pattern**, and this guide covers an orchestrated Saga implemented with **Temporal** — a durable workflow engine that makes writing distributed business logic feel like writing ordinary sequential code.

---

## Why 2PC Fails at Scale

Two-Phase Commit coordinates a transaction across multiple resource managers using a central coordinator. In phase 1 (prepare), all participants vote to commit or abort. In phase 2 (commit/rollback), the coordinator executes the decision.

The problems:

**Blocking**: Participants hold locks while waiting for phase 2. If the coordinator crashes between phases, participants are stuck holding locks indefinitely.

**Tight coupling**: Every participating service must implement the 2PC protocol. This is straightforward for databases — JDBC and XA support it — but not for HTTP services or message brokers.

**Availability**: If any participant is unavailable, the transaction blocks. In a microservices architecture where each service can independently fail or deploy, this is a near-constant condition.

**Latency**: Two network round-trips plus coordination overhead on every transaction.

For a monolith with a single database, transactions work fine. For distributed systems, they're a reliability liability.

---

## The Saga Pattern

A Saga is a sequence of local transactions, each of which updates one service's data and publishes an event or message. If a step fails, compensating transactions are executed in reverse to undo the completed steps.

Two coordination styles:

**Choreography**: Each service listens to events and decides what to do next. No central coordinator. Services are fully decoupled but the overall flow is implicit — to understand what happens during an order, you have to trace events across many services.

**Orchestration**: A central workflow defines the sequence of steps explicitly. The orchestrator calls each service in turn, handles failures, and triggers compensations. The flow is explicit and visible in one place.

Orchestration is generally preferable for complex business flows with non-trivial compensation logic. It's easier to reason about, test, and debug. The tradeoff is that the orchestrator becomes a coordination point — if it goes down, in-flight sagas pause until it recovers.

This codebase implements an orchestrated saga. Temporal is the orchestrator.

---

## Why Temporal Instead of Rolling Your Own

An orchestrated saga needs:

- State persistence: track which steps have completed
- Retry logic: retry failed activities automatically
- Timeout handling: what happens if a service is slow?
- Compensation triggering: run rollbacks when a step fails
- Visibility: show in-flight workflows and their history

You could implement this with a database table tracking saga state and a scheduler polling for timeouts. Teams have done this. It's brittle, operationally complex, and always underfeatured.

Temporal provides all of this with a key architectural property: **durable execution**. A Temporal workflow is a function that persists its execution state after every step. If the worker process crashes mid-workflow, the workflow resumes exactly where it left off when the worker restarts — as if the crash never happened. No special restart logic, no polling tables, no lost state.

---

## The Workflow Interface

```kotlin
@WorkflowInterface
interface OrderSagaWorkflow {
    @WorkflowMethod
    fun execute(order: OrderRequest): OrderResult
}
```

This is the public contract. Clients call `execute()` as if it were a function — they don't need to know about task queues, activity retries, or workflow state. Temporal handles the distributed execution transparently.

The `OrderResult` includes compensation metadata:

```kotlin
enum class SagaStatus { COMPLETED, COMPENSATED }

data class OrderResult(
    val orderId: String,
    val status: SagaStatus,
    val paymentId: String? = null,
    val reservationId: String? = null,
    val shipmentId: String? = null,
    val failureReason: String? = null,
)
```

`COMPENSATED` means the saga failed and compensations ran. `COMPLETED` means all three steps succeeded. The IDs carry through the result so callers can correlate with downstream service records.

---

## The Workflow Implementation: Sequential Code for a Distributed Flow

```kotlin
class OrderSagaWorkflowImpl : OrderSagaWorkflow {

    private val activities: OrderActivities = Workflow.newActivityStub(
        OrderActivities::class.java,
        ActivityOptions.newBuilder()
            .setStartToCloseTimeout(Duration.ofSeconds(10))
            .setScheduleToCloseTimeout(Duration.ofMinutes(2))
            .build(),
    )

    override fun execute(order: OrderRequest): OrderResult {
        // Forward step 1: charge payment
        val paymentId = try {
            activities.chargePayment(order.orderId, order.customerId, order.totalCents)
        } catch (e: Exception) {
            return OrderResult.failed(order.orderId, "Payment failed: ${e.message}")
        }

        // Forward step 2: reserve inventory
        val reservationId = try {
            activities.reserveInventory(order.orderId, order.items)
        } catch (e: Exception) {
            runCompensation("refundPayment") {
                activities.refundPayment(paymentId, order.orderId)
            }
            return OrderResult.failed(order.orderId, "Inventory reservation failed: ${e.message}")
        }

        // Forward step 3: schedule shipment
        val shipmentId = try {
            activities.scheduleShipment(order.orderId, order.customerId, reservationId)
        } catch (e: Exception) {
            runCompensation("releaseInventory") {
                activities.releaseInventory(reservationId, order.orderId)
            }
            runCompensation("refundPayment") {
                activities.refundPayment(paymentId, order.orderId)
            }
            return OrderResult.failed(order.orderId, "Shipment scheduling failed: ${e.message}")
        }

        return OrderResult.success(order.orderId, paymentId, reservationId, shipmentId)
    }
}
```

This looks like ordinary sequential Kotlin — try/catch, function calls, return values. But `activities.chargePayment()` is a call to a remote service that might run on a different machine, take several seconds, and be retried multiple times. Temporal's SDK translates each activity call into a durable checkpoint.

The compensation ordering is explicit and correct: reverse order. When shipment fails, release inventory first (undoing step 2), then refund payment (undoing step 1). The `runCompensation` helper ensures one failed compensation doesn't prevent subsequent ones from running.

---

## Activities: The Actual Service Calls

```kotlin
@ActivityInterface
interface OrderActivities {
    @ActivityMethod fun chargePayment(orderId: String, customerId: String, amountCents: Long): String
    @ActivityMethod fun reserveInventory(orderId: String, items: List<OrderItem>): String
    @ActivityMethod fun scheduleShipment(orderId: String, customerId: String, reservationId: String): String
    @ActivityMethod fun refundPayment(paymentId: String, orderId: String)
    @ActivityMethod fun releaseInventory(reservationId: String, orderId: String)
}
```

Activities are the leaf nodes of the saga — the actual HTTP or gRPC calls to downstream services. Each is annotated with `@ActivityMethod` so Temporal knows to route execution through its task queue system.

The implementation shows production patterns:

```kotlin
override fun chargePayment(orderId: String, customerId: String, amountCents: Long): String {
    val ctx = Activity.getExecutionContext()
    val idempotencyKey = "${ctx.info.workflowId}:chargePayment"

    ctx.heartbeat("initiating charge")
    // ...
}
```

**Idempotency key**: Temporal retries activities on worker failure. The idempotency key (`workflowId:activityType`) lets the downstream payment service deduplicate retried calls. Without this, a network blip that drops the response would trigger a retry and potentially double-charge the customer.

**Heartbeating**: `ctx.heartbeat()` tells Temporal the activity is still alive. For long-running activities (payment processing, file upload), heartbeating prevents Temporal from timing out the activity and retrying it prematurely. Set `heartbeatTimeout` in the activity options when using this.

**Failure injection**: `failOnActivity` lets the demo trigger compensation without external service setup — just point the failure at any step and watch the compensation chain execute.

---

## Compensation Resilience

```kotlin
private fun runCompensation(name: String, action: () -> Unit) {
    try {
        action()
        log.info("Compensation '{}' succeeded", name)
    } catch (e: Exception) {
        log.error("Compensation '{}' FAILED — requires manual review: {}", name, e.message)
    }
}
```

Compensations can fail too. If the refund service is down when we try to compensate, `runCompensation` logs the failure but does not rethrow — ensuring subsequent compensations still run. A failed compensation is a serious operational concern that requires human intervention or a dedicated dead-letter workflow. Alert on these.

---

## Testing with TestWorkflowEnvironment

```kotlin
@Test
fun `scheduleShipment failure triggers both releaseInventory and refundPayment compensations`() {
    stub.chargePaymentResult = "pay-003"
    stub.reserveInventoryResult = "res-003"
    stub.scheduleShipmentError = RuntimeException("Courier API timeout")

    val result = newWorkflowStub().execute(testOrder)

    assertThat(result.status).isEqualTo(SagaStatus.COMPENSATED)
    assertThat(stub.releaseCallCount.get()).isEqualTo(1)
    assertThat(stub.refundCallCount.get()).isEqualTo(1)
    assertThat(stub.releaseReservationIdCapture).isEqualTo("res-003")
    assertThat(stub.refundPaymentIdCapture).isEqualTo("pay-003")
}
```

`TestWorkflowEnvironment` runs the full workflow in-process — no actual Temporal server needed. The `StubActivities` class manually implements `OrderActivities` as a configurable stub. MockK cannot be used directly because its ByteBuddy proxy inherits `@ActivityMethod` annotations onto the concrete class, which Temporal rejects. A plain implementing class sidesteps this.

Each compensation scenario has a dedicated test: failure at step 1 (no compensation needed), step 2 (refund only), step 3 (release + refund), and compensation failure resilience (one failed compensation doesn't block others). This coverage is what gives you confidence to run the saga in production.

---

## The Worker: Connecting to Temporal

```kotlin
fun main() {
    val stubs = WorkflowServiceStubs.newLocalServiceStubs()
    val client = WorkflowClient.newInstance(stubs)
    val factory = WorkerFactory.newInstance(client)

    val worker: Worker = factory.newWorker(TASK_QUEUE)
    worker.registerWorkflowImplementationTypes(OrderSagaWorkflowImpl::class.java)
    worker.registerActivitiesImplementations(OrderActivitiesImpl())

    factory.start()
    log.info("Open http://localhost:8233 to see workflow executions in the Temporal Web UI")
}
```

Workers poll the Temporal server for tasks on a named task queue (`order-saga-queue`). Multiple worker processes can poll the same queue, giving horizontal scalability — just run more workers. The Temporal server persists workflow state; workers are stateless and replaceable.

The Temporal Web UI at `http://localhost:8233` shows every workflow execution: its history, current state, input, output, and any failures. This visibility is one of Temporal's strongest operational advantages.

---

## Key Takeaways

- 2PC fails in microservices: blocking locks, tight coupling, availability requirements, latency overhead
- The Saga pattern replaces atomic transactions with a sequence of local transactions and compensating actions
- Orchestrated sagas make the flow explicit in one place; choreographed sagas keep services decoupled but make the flow implicit
- Temporal provides durable execution: workflows survive worker crashes and resume from where they stopped, without custom restart logic
- The workflow code looks sequential but is actually distributed — Temporal checkpoints after each activity call
- Design compensations first, before the happy path — they're harder to add later and the failure scenarios are where bugs hide
- Activities must be idempotent: Temporal retries on worker failure; derive idempotency keys from `workflowId:activityType`
- Heartbeat long-running activities to prevent premature timeout and retry
- Failed compensations require human intervention — alert on them; log the failed compensation and the IDs needed to fix it manually
- `TestWorkflowEnvironment` runs workflows in-process with no server dependency — test every compensation scenario before shipping
