package com.patterns.eventdriven

import com.fasterxml.jackson.module.kotlin.jacksonObjectMapper
import org.slf4j.LoggerFactory
import java.util.UUID

private val objectMapper = jacksonObjectMapper().apply {
    findAndRegisterModules()
}

/**
 * Transactional outbox pattern implementation.
 *
 * The outbox pattern solves the dual-write problem:
 *   Problem:  DB write and Kafka publish are not atomic. If Kafka is down
 *             after the DB commit, the event is lost. If we publish to Kafka
 *             first and the DB write fails, we've published a phantom event.
 *
 *   Solution: Write the business entity AND the event to the outbox table in the
 *             same DB transaction. A separate relay process (Debezium CDC, or the
 *             [OutboxRelay] coroutine in this demo) picks up unpublished events
 *             and publishes them to Kafka asynchronously.
 *
 * In production with Spring:
 *   @Transactional
 *   fun createOrder(cmd: CreateOrderCommand): Order { ... }
 *
 * In this standalone demo, [AtomicOperation.execute] simulates the transaction
 * boundary — both saves either succeed or both fail (simulated).
 */
class OrderService(
    private val orderRepository: OrderRepository,
    private val outboxRepository: OutboxRepository,
) {
    private val log = LoggerFactory.getLogger(OrderService::class.java)

    /**
     * Creates an order and writes an [OutboxEvent] in the same "transaction".
     *
     * The event payload is keyed by [Order.id] (aggregateId), which becomes the
     * Kafka topic key. This guarantees that all events for a given order go to
     * the same partition → per-order event ordering is preserved.
     */
    fun createOrder(
        customerId: String,
        items: List<OrderItem>,
    ): Order {
        require(customerId.isNotBlank()) { "customerId must not be blank" }
        require(items.isNotEmpty()) { "Order must have at least one item" }

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

        // === Simulated transaction boundary ===
        // In production: @Transactional ensures both saves are atomic.
        // Either both succeed (commit) or both are rolled back on exception.
        simulateTransaction {
            orderRepository.save(order)
            outboxRepository.save(outboxEvent)
        }
        // ======================================

        log.info(
            "[OrderService] Order created id={} customerId={} totalCents={} outboxEventId={}",
            order.id,
            order.customerId,
            order.totalCents,
            outboxEvent.id,
        )

        return order
    }

    /**
     * Simulates a transaction boundary. Both operations are called together.
     * In production this is replaced by Spring's @Transactional or explicit
     * transaction management (e.g., JOOQ DSL.using(config).transaction { ... }).
     */
    private fun simulateTransaction(block: () -> Unit) {
        // In a real implementation, if block() throws, the DB transaction rolls back
        // and neither the order nor the outbox event is persisted.
        block()
    }
}
