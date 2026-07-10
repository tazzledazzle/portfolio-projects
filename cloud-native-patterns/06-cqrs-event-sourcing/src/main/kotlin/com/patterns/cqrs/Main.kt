package com.patterns.cqrs

import com.patterns.cqrs.commands.DeductInventoryCommand
import com.patterns.cqrs.commands.InitializeInventoryCommand
import com.patterns.cqrs.commands.RestockInventoryCommand
import com.patterns.cqrs.projection.InventoryProjection

fun main() {
    println("=== CQRS + Event Sourcing Demo ===")
    println("No framework (Axon) — hand-rolled to show mechanics")
    println()

    val eventStore = EventStore(snapshotFrequency = 5)  // snapshot every 5 events for demo
    val projection = InventoryProjection()
    val bus = CommandBus(eventStore, projection)

    val widgetSku = "SKU-WIDGET"
    val gadgetSku = "SKU-GADGET"

    // ─── Initialize Inventory ─────────────────────────────────────────────────
    println("--- Initialize inventory ---")
    bus.dispatch(InitializeInventoryCommand(widgetSku, widgetSku, initialQuantity = 100, reorderThreshold = 15))
    bus.dispatch(InitializeInventoryCommand(gadgetSku, gadgetSku, initialQuantity = 50, reorderThreshold = 10))
    printProjection(projection)

    // ─── Deduct Inventory (normal) ────────────────────────────────────────────
    println("--- Deduct 30 widgets (3 orders of 10) ---")
    repeat(3) { i ->
        bus.dispatch(DeductInventoryCommand(widgetSku, widgetSku, quantity = 10, reason = "order-${i + 1}"))
    }
    printProjection(projection)

    // ─── Deduct below reorder threshold ──────────────────────────────────────
    println("--- Deduct 60 widgets (triggers reorder threshold event) ---")
    bus.dispatch(DeductInventoryCommand(widgetSku, widgetSku, quantity = 60, reason = "bulk-order"))
    printProjection(projection)

    // ─── Restock ─────────────────────────────────────────────────────────────
    println("--- Restock 200 widgets from supplier ---")
    bus.dispatch(RestockInventoryCommand(widgetSku, widgetSku, quantity = 200, supplierReference = "PO-2026-001"))
    printProjection(projection)

    // ─── Event Log (the source of truth) ─────────────────────────────────────
    println("--- Event log for $widgetSku ---")
    eventStore.getEventLog(widgetSku).forEach { event ->
        println("  [seq=${event.sequenceNumber}] ${event::class.simpleName} at ${event.occurredAt}")
    }
    println()

    // ─── Snapshot state ───────────────────────────────────────────────────────
    val snapshot = eventStore.getSnapshot(widgetSku)
    if (snapshot != null) {
        println("--- Snapshot for $widgetSku ---")
        println("  Snapshot at seq=${snapshot.snapshotAtSequence}: qty=${snapshot.quantity}")
        println("  On next load, only events after seq=${snapshot.snapshotAtSequence} will be replayed")
        println()
    }

    // ─── Reload aggregate from event store (prove replay works) ───────────────
    println("--- Reloading aggregate from event store (replay from snapshot + delta) ---")
    val reloaded = eventStore.load(widgetSku)
    println("  Reloaded state: $reloaded")
    println()

    // ─── Summary ─────────────────────────────────────────────────────────────
    println("=== Summary ===")
    println("Aggregates tracked  : ${eventStore.aggregateCount()}")
    println("Total events stored : ${eventStore.totalEventCount()}")
    println("Total SKUs in view  : ${projection.getTotalSkuCount()}")
    println("Low stock SKUs      : ${projection.getLowStockSkus().map { it.sku }}")
}

private fun printProjection(projection: InventoryProjection) {
    println("  [Projection] Current inventory view:")
    projection.getAllInventory().values.forEach { view ->
        val flag = if (view.belowReorderThreshold) " ⚠ LOW STOCK" else ""
        println("    ${view.sku}: ${view.currentQuantity} units (threshold=${view.reorderThreshold})$flag")
    }
    println()
}
