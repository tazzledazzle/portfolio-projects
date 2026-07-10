package com.patterns.eventdriven

import java.time.Instant

enum class OrderStatus { PENDING, CONFIRMED, CANCELLED }

data class OrderItem(
    val sku: String,
    val quantity: Int,
    val unitPriceCents: Long,
)

data class Order(
    val id: String,
    val customerId: String,
    val items: List<OrderItem>,
    val status: OrderStatus = OrderStatus.PENDING,
    val totalCents: Long = items.sumOf { it.unitPriceCents * it.quantity },
    val createdAt: Instant = Instant.now(),
)

/** Domain event published when an order is created. */
data class OrderCreatedEvent(
    val orderId: String,
    val customerId: String,
    val totalCents: Long,
    val itemCount: Int,
    val occurredAt: Instant = Instant.now(),
)
