package com.patterns.eventdriven

import kotlinx.coroutines.delay
import kotlinx.coroutines.runBlocking

/**
 * Demonstrates the transactional outbox pattern end-to-end:
 *
 * 1. [OrderService.createOrder] writes the order + outbox event atomically.
 * 2. [OutboxRelay] polls for unpublished events and "publishes" to Kafka.
 * 3. If a publish fails, the event stays in the outbox — no event loss.
 * 4. Successful publishes are marked so they're not republished.
 */
fun main() = runBlocking {
    println("=== Transactional Outbox Pattern Demo ===")
    println()

    val orderRepository = OrderRepository()
    val outboxRepository = OutboxRepository()
    val orderService = OrderService(orderRepository, outboxRepository)

    // simulateFailureEveryN=3: every 3rd Kafka publish attempt will fail.
    // This demonstrates that events remain in the outbox and are retried.
    val relay = OutboxRelay(
        outboxRepository = outboxRepository,
        pollIntervalMs = 500L,
        simulateFailureEveryN = 3,
    )

    // Start the relay before creating orders (simulates Debezium running in background)
    relay.start()

    println("--- Creating 5 orders ---")
    println()

    val customers = listOf("cust-alice", "cust-bob", "cust-carol")
    val items = listOf(
        OrderItem("SKU-WIDGET", 2, 999L),
        OrderItem("SKU-GADGET", 1, 4999L),
    )

    repeat(5) { i ->
        val order = orderService.createOrder(
            customerId = customers[i % customers.size],
            items = items,
        )
        println("[main] Created order ${i + 1}/5: id=${order.id} total=${order.totalCents}¢")
        delay(200L)  // Stagger creation to show relay picking them up incrementally
    }

    println()
    println("--- Waiting for relay to process all events (some will fail and retry) ---")
    println()

    // Give the relay enough time to process all events including retries.
    // With simulateFailureEveryN=3, the relay may need 2-3 poll cycles to clear the outbox.
    delay(4_000L)

    println()
    println("=== Final State ===")
    println("Orders created      : ${orderRepository.count()}")
    println("Outbox total events : ${outboxRepository.findAll().size}")
    println("Published           : ${outboxRepository.countPublished()}")
    println("Still pending       : ${outboxRepository.countUnpublished()}")
    println()

    if (outboxRepository.countUnpublished() == 0) {
        println("SUCCESS: All events published. Outbox is clear.")
    } else {
        println("PENDING: ${outboxRepository.countUnpublished()} event(s) still in outbox.")
        println("         The relay will continue retrying until they succeed.")
    }

    relay.stop()
    println()
    println("Demo complete.")
}
