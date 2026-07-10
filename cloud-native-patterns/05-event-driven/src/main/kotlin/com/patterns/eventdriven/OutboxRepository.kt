package com.patterns.eventdriven

import java.time.Instant
import java.util.concurrent.ConcurrentHashMap

/**
 * In-memory outbox repository.
 *
 * In production: a table in the same PostgreSQL database as the business entities.
 * The outbox pattern's guarantee comes from writing to this table in the same
 * transaction as the business write — they either both commit or both roll back.
 */
class OutboxRepository {
    private val store = ConcurrentHashMap<String, OutboxEvent>()

    fun save(event: OutboxEvent): OutboxEvent {
        store[event.id] = event
        return event
    }

    /** Returns all events that have not yet been successfully published. */
    fun findUnpublished(): List<OutboxEvent> =
        store.values.filter { !it.published }.sortedBy { it.createdAt }

    /** Marks an event as successfully published. */
    fun markPublished(eventId: String): OutboxEvent? {
        val event = store[eventId] ?: return null
        val updated = event.copy(
            published = true,
            publishedAt = Instant.now(),
        )
        store[eventId] = updated
        return updated
    }

    /** Records a failed publish attempt (for retry logic and dead-letter monitoring). */
    fun incrementAttempts(eventId: String): OutboxEvent? {
        val event = store[eventId] ?: return null
        val updated = event.copy(publishAttempts = event.publishAttempts + 1)
        store[eventId] = updated
        return updated
    }

    fun findAll(): List<OutboxEvent> = store.values.toList()

    fun countUnpublished(): Int = store.values.count { !it.published }

    fun countPublished(): Int = store.values.count { it.published }
}
