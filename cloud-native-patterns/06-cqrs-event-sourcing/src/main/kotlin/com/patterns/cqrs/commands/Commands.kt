package com.patterns.cqrs.commands

/**
 * Command marker interface.
 *
 * Commands are imperative intentions — they may be rejected.
 * Contrast with events, which are immutable facts that already happened.
 */
sealed interface Command {
    val aggregateId: String
}

/**
 * Deduct [quantity] units of [sku] from inventory.
 * Rejected if current stock < quantity.
 */
data class DeductInventoryCommand(
    override val aggregateId: String,  // SKU is the aggregate ID for inventory
    val sku: String,
    val quantity: Int,
    val reason: String = "order",
) : Command

/**
 * Restock [quantity] units of [sku].
 */
data class RestockInventoryCommand(
    override val aggregateId: String,
    val sku: String,
    val quantity: Int,
    val supplierReference: String,
) : Command

/**
 * Initialize inventory for a new SKU with [initialQuantity] units.
 */
data class InitializeInventoryCommand(
    override val aggregateId: String,
    val sku: String,
    val initialQuantity: Int,
    val reorderThreshold: Int = 10,
) : Command
