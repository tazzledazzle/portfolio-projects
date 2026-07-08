package com.patterns.eventdriven

import java.util.concurrent.ConcurrentHashMap

/**
 * In-memory order repository.
 *
 * In production: a JPA/JOOQ repository backed by PostgreSQL.
 * The critical invariant for the outbox pattern is that [save] and
 * [OutboxRepository.save] are called within the same database transaction.
 */
class OrderRepository {
    private val store = ConcurrentHashMap<String, Order>()

    fun save(order: Order): Order {
        store[order.id] = order
        return order
    }

    fun findById(id: String): Order? = store[id]

    fun findAll(): List<Order> = store.values.toList()

    fun count(): Int = store.size
}
