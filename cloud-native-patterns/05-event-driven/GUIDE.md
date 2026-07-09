# Event-Driven Architecture and the Transactional Outbox Pattern

Event-driven architecture decouples services by having them communicate through events rather than direct calls. Instead of `orders-service` calling `inventory-service` and `notification-service` synchronously, it publishes an `OrderCreated` event to a message broker, and those services consume it independently. This gives you loose coupling, independent scaling, and the ability to add new consumers without touching the producer.

But there's a catch that trips up most teams the first time: **the dual-write problem**. This guide covers what it is, why both obvious solutions fail, and how the Transactional Outbox pattern solves it — using the real Kotlin implementation in this codebase.

---

## Why Event-Driven is Worth the Complexity

Before getting into the failure modes, it's worth being clear on what you're buying:

**Loose coupling**: `orders-service` doesn't know or care about `inventory-service`. It just publishes an event. If the inventory service goes down, orders keep being placed. When inventory comes back up, it catches up from the broker.

**Independent scaling**: A high-volume sales event might cause `notification-service` to lag behind. That lag doesn't propagate upstream — orders still process normally, and notifications catch up when capacity allows.

**Fan-out without coordination**: Adding a new downstream consumer (say, an analytics pipeline) means subscribing to the existing event stream. The producer doesn't need to be modified.

These benefits are real but they come with a consistency model you have to reason about carefully: **eventual consistency**. Events are processed asynchronously, so consumers may be seconds or minutes behind the producer. That's acceptable for notifications and analytics. It's not acceptable for inventory reservation that must happen before shipment.

---

## The Dual-Write Problem

The naive implementation of event publishing looks like this:

```kotlin
// Option A: DB write first
val order = orderRepository.save(order)
kafkaProducer.publish("OrderCreated", order)  // What if this fails?

// Option B: Kafka publish first
kafkaProducer.publish("OrderCreated", order)  // What if DB write fails?
val order = orderRepository.save(order)
```

Both orderings have failure modes:

**Write to DB, then publish to Kafka**: The DB commit succeeds, the application crashes before publishing, or Kafka is temporarily unavailable. The order exists in the DB but no event was published. Downstream services never process it. Silent data loss.

**Publish to Kafka, then write to DB**: The event is published, then the DB write fails. Now downstream services are processing an event for an order that doesn't exist. Phantom processing, potentially including charging a customer for an order that failed.

The root cause: writing to a database and publishing to a message broker are two separate systems. There's no atomic commit spanning both without distributed transactions, and distributed transactions have their own serious problems at scale.

---

## The Transactional Outbox Pattern

The outbox pattern eliminates the dual-write problem by turning two writes into one:

1. Write the business entity (the order) to the database
2. Write the event to an **outbox table** in the **same database**
3. Do both in a **single database transaction**

A separate relay process polls the outbox table and publishes pending events to Kafka. The database transaction is the atomicity boundary — either both the order and the event record are saved, or neither is.

```kotlin
fun createOrder(customerId: String, items: List<OrderItem>): Order {
    val order = Order(
        id = UUID.randomUUID().toString(),
        customerId = customerId,
        items = items,
    )

    val domainEvent = OrderCreatedEvent(
        orderId = order.id,
        customerId = order.customerId,
        totalCents = order.totalCents,
        itemCount = order.items.size,
    )

    val outboxEvent = OutboxEvent(
        aggregateType = "Order",
        aggregateId = order.id,   // Kafka topic key — ensures per-order ordering
        eventType = "OrderCreated",
        payload = objectMapper.writeValueAsString(domainEvent),
    )

    // Both saves happen inside the same DB transaction
    simulateTransaction {
        orderRepository.save(order)
        outboxRepository.save(outboxEvent)
    }

    return order
}
```

In production with Spring, `@Transactional` ensures the atomicity. In this demo, `simulateTransaction { }` makes the intent explicit. The critical invariant: **if the order is saved, the outbox event is saved; if the transaction rolls back, both are rolled back**.

---

## The OutboxEvent Schema

```kotlin
data class OutboxEvent(
    val id: String = UUID.randomUUID().toString(),
    val aggregateType: String,     // "Order"
    val aggregateId: String,       // order.id — used as Kafka topic key
    val eventType: String,         // "OrderCreated"
    val payload: String,           // JSON-serialized event
    val createdAt: Instant = Instant.now(),
    val published: Boolean = false,
    val publishedAt: Instant? = null,
    val publishAttempts: Int = 0,
)
```

`aggregateId` becomes the Kafka message key. Kafka partitions messages by key — all events with the same key go to the same partition, in order. This guarantees that all events for a given order (created, updated, cancelled) are processed in the order they were written, regardless of how many partitions the topic has.

`published` and `publishAttempts` are for the polling relay. The CDC approach (Debezium) doesn't need them because it reads the WAL log directly — but for the in-memory simulation, explicit tracking is clearer.

---

## The OutboxRelay: Bridging Database to Kafka

The relay is responsible for picking up unpublished events and delivering them to Kafka:

```kotlin
suspend fun poll() {
    val unpublished = outboxRepository.findUnpublished()
    if (unpublished.isEmpty()) return

    for (event in unpublished) {
        try {
            outboxRepository.incrementAttempts(event.id)
            publishToKafka(event)
            outboxRepository.markPublished(event.id)
        } catch (e: Exception) {
            // Publish failed — leave the event unpublished for next poll cycle.
            // This is the core of the outbox guarantee: we never lose the event.
            log.warn("[OutboxRelay] Publish failed for event id={} attempt={}: {}",
                event.id, event.publishAttempts + 1, e.message)
        }
    }
}
```

The `catch` block is the most important part of this implementation. When `publishToKafka` throws — because Kafka is down, the network is partitioned, or any other reason — the event is left in the outbox with `published=false`. On the next poll cycle, it will be retried. **Events accumulate in the outbox when Kafka is unavailable, and drain when Kafka recovers.**

The simulated Kafka publish shows what the production call looks like:

```kotlin
private fun publishToKafka(event: OutboxEvent) {
    val topicName = "orders.${event.aggregateType.lowercase()}.${event.eventType.lowercase()}"
    val messageKey = event.aggregateId  // Per-entity partition key

    // Production: KafkaProducer.send(ProducerRecord(topicName, messageKey, payload))
}
```

The topic name convention mirrors Debezium's format: `<connector>.<schema>.<table>`. This makes it easy to correlate outbox events with CDC events when debugging.

---

## At-Least-Once Delivery and Idempotent Consumers

The relay provides **at-least-once delivery**, not exactly-once. Here's why:

1. Relay polls and finds unpublished events
2. Relay calls `publishToKafka(event)` — Kafka accepts and persists the message
3. Relay crashes before `markPublished(event.id)` is called
4. Relay restarts and finds the same event still marked `published=false`
5. Relay publishes the same event again

The consumer receives the same event twice. This is expected behavior — not a bug. The fix is to make consumers **idempotent**: processing the same event twice should produce the same result as processing it once.

Common deduplication approaches:
- Store `event.id` in a `processed_events` table with a unique constraint; skip events already in the table
- Use the `aggregateId` + sequence number as a cursor; reject events with sequence numbers already seen
- For side effects that are naturally idempotent (setting a status to "confirmed"), deduplication may not be needed

The `OutboxEvent.id` UUID is the deduplication key. Consumers that need exactly-once semantics should check this key before processing.

---

## Polling vs. CDC: Two Relay Approaches

This implementation uses a **polling relay** — a loop that queries the outbox table on an interval. Simple to implement, works with any database, no additional infrastructure required.

In production, the preferred approach is **Change Data Capture (CDC) with Debezium**. Debezium monitors the PostgreSQL Write-Ahead Log (WAL), captures every INSERT to the outbox table, and publishes it to Kafka — no polling needed. Benefits:

- Near-real-time event delivery (milliseconds vs. poll interval)
- No load on the primary database from polling queries
- Exactly-once delivery from outbox to Kafka (WAL offset = position)
- Automatic handling of large event volumes without tuning poll frequency

The tradeoff: Debezium is another system to operate. For teams starting out, the polling relay is a valid and simpler approach. The application code is identical — swap the relay implementation without touching `OrderService`.

---

## Tradeoffs

**Latency**: Events aren't published until the relay runs. Poll interval adds latency to event delivery. CDC reduces this to near-zero but adds infrastructure complexity.

**Outbox table growth**: Outbox records accumulate. You need a cleanup job to delete old published events. Set a retention window appropriate to your debugging needs (7 days is common).

**No global ordering**: Events for different orders can arrive at Kafka in any order. Only events for the *same* order (same `aggregateId` = same Kafka partition) are ordered. Consumers that need cross-aggregate ordering need a different approach.

**Schema coupling**: The event payload schema couples producer and consumer. Changes to `OrderCreatedEvent` must be backward-compatible or versioned. Use an Avro or Protobuf schema registry to enforce compatibility.

**When to use something simpler**: If you have synchronous SLA requirements, use synchronous calls. If you have exactly two services communicating, direct HTTP is simpler. The outbox pattern pays for itself when you have multiple consumers, need to decouple producer and consumer availability, or are integrating with an event streaming platform.

---

## Key Takeaways

- Event-driven architecture decouples services at the cost of eventual consistency — acceptable for notifications and analytics, not for synchronous workflows
- The dual-write problem: no atomic commit spans a database and a message broker — both orderings of write+publish have failure modes that lose or phantom events
- The Transactional Outbox pattern: write the event to the same database table in the same transaction as the business entity — atomicity guaranteed by the DB transaction
- A relay process (polling or CDC via Debezium) picks up unpublished events and delivers them to Kafka
- The `catch` block in the relay is the durability guarantee — failed publishes stay in the outbox and are retried
- Delivery is at-least-once; consumers must be idempotent, using `event.id` as the deduplication key
- `aggregateId` as the Kafka message key routes all events for the same entity to the same partition, preserving per-entity ordering
- Debezium CDC is the production relay for low-latency delivery; polling is simpler to operate for lower-volume workloads
- Plan outbox table cleanup from day one — published events must be deleted or the table grows unboundedly
