package com.patterns.cqrs.events

import java.time.Instant

/**
 * Domain event marker interface.
 *
 * Events are immutable facts — they record what happened, past tense.
 * Naming convention: past participle ("InventoryDeducted", not "DeductInventory").
 *
 * Each event carries the aggregate ID and a monotonically increasing sequence number
 * (assigned by the EventStore) so the full history can be reconstructed in order.
 */
sealed interface DomainEvent {
    val aggregateId: String
    val occurredAt: Instant
    var sequenceNumber: Long  // Set by EventStore on append; 0 until stored
}

data class InventoryInitialized(
    override val aggregateId: String,
    val sku: String,
    val initialQuantity: Int,
    val reorderThreshold: Int,
    override val occurredAt: Instant = Instant.now(),
    override var sequenceNumber: Long = 0,
) : DomainEvent

data class InventoryDeducted(
    override val aggregateId: String,
    val sku: String,
    val quantity: Int,
    val remainingQuantity: Int,
    val reason: String,
    override val occurredAt: Instant = Instant.now(),
    override var sequenceNumber: Long = 0,
) : DomainEvent

data class InventoryRestocked(
    override val aggregateId: String,
    val sku: String,
    val quantity: Int,
    val newQuantity: Int,
    val supplierReference: String,
    override val occurredAt: Instant = Instant.now(),
    override var sequenceNumber: Long = 0,
) : DomainEvent

/**
 * Published when stock falls below [reorderThreshold] after a deduction.
 * Consumers (purchasing system) react to this to place a supplier order.
 */
data class InventoryBelowReorderThreshold(
    override val aggregateId: String,
    val sku: String,
    val currentQuantity: Int,
    val reorderThreshold: Int,
    override val occurredAt: Instant = Instant.now(),
    override var sequenceNumber: Long = 0,
) : DomainEvent
