package com.patterns.cqrs.aggregate

import com.patterns.cqrs.commands.DeductInventoryCommand
import com.patterns.cqrs.commands.InitializeInventoryCommand
import com.patterns.cqrs.commands.RestockInventoryCommand
import com.patterns.cqrs.events.DomainEvent
import com.patterns.cqrs.events.InventoryBelowReorderThreshold
import com.patterns.cqrs.events.InventoryDeducted
import com.patterns.cqrs.events.InventoryInitialized
import com.patterns.cqrs.events.InventoryRestocked

/**
 * Snapshot of aggregate state at a point in time.
 *
 * Used to avoid replaying the full event history on every load.
 * Best practice: snapshot every N events (N=10 here). On load, fetch the
 * latest snapshot, then replay only the events after it.
 *
 * In production, snapshots are stored in EventStoreDB or a separate table.
 */
data class InventorySnapshot(
    val aggregateId: String,
    val sku: String,
    val quantity: Int,
    val reorderThreshold: Int,
    val initialized: Boolean,
    val snapshotAtSequence: Long,  // The event sequence number this snapshot was taken after
)

/**
 * Inventory aggregate — the write-side domain object.
 *
 * The aggregate enforces invariants (e.g., can't deduct below zero) and
 * produces events that are the source of truth. The current state is always
 * derivable by replaying events from the beginning (or from a snapshot).
 *
 * Pattern: command handler validates and produces events → apply() updates state.
 * The apply() methods are also used when replaying from the event store.
 *
 * This class is NEVER used for reads — that's what [com.patterns.cqrs.projection.InventoryProjection] is for.
 */
class InventoryAggregate(val aggregateId: String) {
    var sku: String = ""
        private set
    var quantity: Int = 0
        private set
    var reorderThreshold: Int = 10
        private set
    var initialized: Boolean = false
        private set

    // Pending events produced by command handlers — the EventStore drains this list.
    private val pendingEvents = mutableListOf<DomainEvent>()

    fun drainPendingEvents(): List<DomainEvent> {
        val events = pendingEvents.toList()
        pendingEvents.clear()
        return events
    }

    // ─── Command Handlers ────────────────────────────────────────────────────

    fun handle(cmd: InitializeInventoryCommand) {
        require(!initialized) { "Inventory for ${cmd.sku} is already initialized" }
        require(cmd.initialQuantity >= 0) { "Initial quantity must be non-negative" }

        val event = InventoryInitialized(
            aggregateId = aggregateId,
            sku = cmd.sku,
            initialQuantity = cmd.initialQuantity,
            reorderThreshold = cmd.reorderThreshold,
        )
        apply(event)
        pendingEvents.add(event)
    }

    fun handle(cmd: DeductInventoryCommand): List<DomainEvent> {
        require(initialized) { "Inventory for ${cmd.sku} is not initialized" }
        require(cmd.quantity > 0) { "Deduction quantity must be positive" }
        check(quantity >= cmd.quantity) {
            "Insufficient inventory: have $quantity units of ${cmd.sku}, requested ${cmd.quantity}"
        }

        val remaining = quantity - cmd.quantity
        val deducted = InventoryDeducted(
            aggregateId = aggregateId,
            sku = sku,
            quantity = cmd.quantity,
            remainingQuantity = remaining,
            reason = cmd.reason,
        )
        apply(deducted)
        pendingEvents.add(deducted)

        // Produce a secondary event if stock falls below threshold.
        // This is a domain event that downstream services can react to.
        if (remaining < reorderThreshold) {
            val belowThreshold = InventoryBelowReorderThreshold(
                aggregateId = aggregateId,
                sku = sku,
                currentQuantity = remaining,
                reorderThreshold = reorderThreshold,
            )
            apply(belowThreshold)
            pendingEvents.add(belowThreshold)
        }

        return drainPendingEvents()
    }

    fun handle(cmd: RestockInventoryCommand) {
        require(initialized) { "Inventory for ${cmd.sku} is not initialized" }
        require(cmd.quantity > 0) { "Restock quantity must be positive" }

        val newQty = quantity + cmd.quantity
        val event = InventoryRestocked(
            aggregateId = aggregateId,
            sku = sku,
            quantity = cmd.quantity,
            newQuantity = newQty,
            supplierReference = cmd.supplierReference,
        )
        apply(event)
        pendingEvents.add(event)
    }

    // ─── Apply (state mutators) ───────────────────────────────────────────────
    // apply() is called both when handling a new command AND when replaying events.
    // It must be side-effect-free (no external calls, no validation).

    fun apply(event: DomainEvent) {
        when (event) {
            is InventoryInitialized -> {
                sku = event.sku
                quantity = event.initialQuantity
                reorderThreshold = event.reorderThreshold
                initialized = true
            }
            is InventoryDeducted -> {
                quantity -= event.quantity
            }
            is InventoryRestocked -> {
                quantity = event.newQuantity
            }
            is InventoryBelowReorderThreshold -> {
                // No state change — this event is informational for downstream consumers.
            }
        }
    }

    /** Restore state from a snapshot, skipping event replay up to [snapshot.snapshotAtSequence]. */
    fun restoreFromSnapshot(snapshot: InventorySnapshot) {
        sku = snapshot.sku
        quantity = snapshot.quantity
        reorderThreshold = snapshot.reorderThreshold
        initialized = snapshot.initialized
    }

    fun toSnapshot(snapshotAtSequence: Long): InventorySnapshot =
        InventorySnapshot(
            aggregateId = aggregateId,
            sku = sku,
            quantity = quantity,
            reorderThreshold = reorderThreshold,
            initialized = initialized,
            snapshotAtSequence = snapshotAtSequence,
        )

    override fun toString(): String =
        "InventoryAggregate(sku=$sku, quantity=$quantity, threshold=$reorderThreshold)"
}
