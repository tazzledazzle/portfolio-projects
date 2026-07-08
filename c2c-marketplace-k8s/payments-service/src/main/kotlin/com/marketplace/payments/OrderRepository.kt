package com.marketplace.payments

import kotlinx.serialization.Serializable
import org.jetbrains.exposed.sql.ResultRow
import org.jetbrains.exposed.sql.Table
import org.jetbrains.exposed.sql.insert
import org.jetbrains.exposed.sql.javatime.timestamp
import org.jetbrains.exposed.sql.selectAll
import org.jetbrains.exposed.sql.transactions.transaction
import org.jetbrains.exposed.sql.update
import java.time.Instant
import java.util.UUID

object OrderTable : Table("orders") {
    val id = varchar("id", 36)
    val listingId = varchar("listing_id", 36)
    val buyerId = varchar("buyer_id", 36)
    val sellerId = varchar("seller_id", 36)
    val amountCents = integer("amount_cents")
    val status = varchar("status", 16)
    override val primaryKey = PrimaryKey(id)
}

object EscrowHoldTable : Table("escrow_holds") {
    val orderId = varchar("order_id", 36).references(OrderTable.id)
    val status = varchar("status", 16)
    val heldAt = timestamp("held_at")
    val releasedAt = timestamp("released_at").nullable()
    override val primaryKey = PrimaryKey(orderId)
}

@Serializable
data class Order(
    val id: String,
    val listingId: String,
    val buyerId: String,
    val sellerId: String,
    val amountCents: Int,
    val status: String
)

@Serializable
data class CreateOrderRequest(
    val listingId: String,
    val buyerId: String,
    val sellerId: String,
    val amountCents: Int
)

/**
 * The one repository in this whole system where two tables MUST commit
 * together (order row + escrow hold row) -- see TDD.md section 3. Both
 * inserts happen inside a single `transaction {}` block; there is no
 * version of this repository that writes the order and then separately,
 * later, writes the hold.
 */
class OrderRepository {

    fun createWithHold(req: CreateOrderRequest): Order {
        val id = UUID.randomUUID().toString()
        val now = Instant.now()
        transaction {
            OrderTable.insert {
                it[OrderTable.id] = id
                it[listingId] = req.listingId
                it[buyerId] = req.buyerId
                it[sellerId] = req.sellerId
                it[amountCents] = req.amountCents
                it[status] = EscrowStatus.HELD.name
            }
            EscrowHoldTable.insert {
                it[orderId] = id
                it[status] = EscrowStatus.HELD.name
                it[heldAt] = now
            }
        }
        return Order(id, req.listingId, req.buyerId, req.sellerId, req.amountCents, EscrowStatus.HELD.name)
    }

    fun findById(id: String): Order? = transaction {
        OrderTable.selectAll().where { OrderTable.id eq id }
            .map { it.toOrder() }
            .singleOrNull()
    }

    fun currentEscrowStatus(orderId: String): EscrowStatus? = transaction {
        EscrowHoldTable.selectAll().where { EscrowHoldTable.orderId eq orderId }
            .map { EscrowStatus.valueOf(it[EscrowHoldTable.status]) }
            .singleOrNull()
    }

    /**
     * Applies an escrow event via [EscrowStateMachine] and persists the
     * result, keeping order.status and escrow_holds.status in lockstep --
     * they're two tables but one logical piece of state, updated in the
     * same transaction every time.
     */
    fun applyEvent(orderId: String, event: EscrowEvent): EscrowStatus = transaction {
        val current = EscrowHoldTable.selectAll().where { EscrowHoldTable.orderId eq orderId }
            .map { EscrowStatus.valueOf(it[EscrowHoldTable.status]) }
            .singleOrNull() ?: throw NoSuchElementException("no escrow hold for order $orderId")

        val next = EscrowStateMachine.transition(current, event)

        EscrowHoldTable.update({ EscrowHoldTable.orderId eq orderId }) {
            it[status] = next.name
            if (next != EscrowStatus.HELD) it[releasedAt] = Instant.now()
        }
        OrderTable.update({ OrderTable.id eq orderId }) {
            it[status] = next.name
        }
        next
    }

    private fun ResultRow.toOrder() = Order(
        id = this[OrderTable.id],
        listingId = this[OrderTable.listingId],
        buyerId = this[OrderTable.buyerId],
        sellerId = this[OrderTable.sellerId],
        amountCents = this[OrderTable.amountCents],
        status = this[OrderTable.status]
    )
}
