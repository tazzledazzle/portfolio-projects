package com.patterns.eventdriven

import java.time.Instant
import java.util.UUID

/**
 * Outbox table row.
 *
 * In production this is a real DB table row (same schema, same transaction as
 * the business entity). Debezium CDC captures inserts from this table and
 * publishes to Kafka. This decouples the write guarantee from the messaging
 * system availability — if Kafka is down, events accumulate in the outbox and
 * the relay retries.
 *
 * Key fields:
 *  - [aggregateId]: used as the Kafka message key → guarantees per-entity ordering.
 *  - [published]: relay sets this to true after successful Kafka publish. The CDC
 *    approach (Debezium) doesn't need this flag because it uses the WAL log offset —
 *    but for this in-memory simulation we track it explicitly.
 */
data class OutboxEvent(
    val id: String = UUID.randomUUID().toString(),
    val aggregateType: String,       // e.g. "Order"
    val aggregateId: String,         // e.g. order.id — used as Kafka topic key
    val eventType: String,           // e.g. "OrderCreated"
    val payload: String,             // JSON-serialized event
    val createdAt: Instant = Instant.now(),
    val published: Boolean = false,
    val publishedAt: Instant? = null,
    val publishAttempts: Int = 0,
)
