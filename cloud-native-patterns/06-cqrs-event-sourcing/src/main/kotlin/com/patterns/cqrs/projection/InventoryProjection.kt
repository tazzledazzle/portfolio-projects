package com.patterns.cqrs.projection

import com.patterns.cqrs.events.DomainEvent
import com.patterns.cqrs.events.InventoryBelowReorderThreshold
import com.patterns.cqrs.events.InventoryDeducted
import com.patterns.cqrs.events.InventoryInitialized
import com.patterns.cqrs.events.InventoryRestocked
import org.slf4j.LoggerFactory

/**
 * Read-side view model for inventory.
 *
 * Updated by replaying events from the [com.patterns.cqrs.EventStore].
 * This is a denormalized, query-optimized representation — not the canonical state.
 * The canonical state is always the event log.
 *
 * Consistency note: projections are EVENTUALLY CONSISTENT. A command that
 * just succeeded may not yet be reflected here. Surface this in your API:
 * "Inventory levels as of [lastUpdatedAt]" rather than presenting them as current.
 */
data class InventoryView(
    val sku: String,
    val currentQuantity: Int,
    val reorderThreshold: Int,
    val belowReorderThreshold: Boolean = currentQuantity < reorderThreshold,
    val lastEventSequence: Long = 0,
)

class InventoryProjection {
    private val log = LoggerFactory.getLogger(InventoryProjection::class.java)
    private val views = mutableMapOf<String, InventoryView>()

    /**
     * Apply an event to update the read model.
     *
     * In production this is called:
     * a) During command processing (synchronous update, same JVM)
     * b) By a Kafka consumer replaying events from the event stream (async)
     * c) During projection rebuild (replay all events from EventStore)
     */
    fun on(event: DomainEvent) {
        when (event) {
            is InventoryInitialized -> {
                views[event.sku] = InventoryView(
                    sku = event.sku,
                    currentQuantity = event.initialQuantity,
                    reorderThreshold = event.reorderThreshold,
                    lastEventSequence = event.sequenceNumber,
                )
                log.debug("[Projection] Initialized {} qty={}", event.sku, event.initialQuantity)
            }
            is InventoryDeducted -> {
                views.compute(event.sku) { _, existing ->
                    existing?.copy(
                        currentQuantity = event.remainingQuantity,
                        lastEventSequence = event.sequenceNumber,
                    )
                }
                log.debug("[Projection] Deducted {} qty={} remaining={}", event.sku, event.quantity, event.remainingQuantity)
            }
            is InventoryRestocked -> {
                views.compute(event.sku) { _, existing ->
                    existing?.copy(
                        currentQuantity = event.newQuantity,
                        lastEventSequence = event.sequenceNumber,
                    )
                }
                log.debug("[Projection] Restocked {} new_qty={}", event.sku, event.newQuantity)
            }
            is InventoryBelowReorderThreshold -> {
                // Projection already reflects the low stock via currentQuantity — this
                // event is primarily for downstream notification (purchasing system, alerts).
                log.warn(
                    "[Projection] LOW STOCK: {} is at {} units (threshold: {})",
                    event.sku,
                    event.currentQuantity,
                    event.reorderThreshold,
                )
            }
        }
    }

    // ─── Query methods (the "Q" in CQRS) ─────────────────────────────────────

    fun getInventory(sku: String): InventoryView? = views[sku]

    fun getAllInventory(): Map<String, InventoryView> = views.toMap()

    fun getLowStockSkus(): List<InventoryView> =
        views.values.filter { it.belowReorderThreshold }.sortedBy { it.currentQuantity }

    fun getTotalSkuCount(): Int = views.size
}
