package com.marketplace.listings

import kotlinx.serialization.Serializable
import org.jetbrains.exposed.sql.ResultRow
import org.jetbrains.exposed.sql.SortOrder
import org.jetbrains.exposed.sql.insert
import org.jetbrains.exposed.sql.select
import org.jetbrains.exposed.sql.selectAll
import org.jetbrains.exposed.sql.transactions.transaction
import org.jetbrains.exposed.sql.update
import java.time.Instant
import java.util.UUID

@Serializable
data class Listing(
    val id: String,
    val sellerId: String,
    val title: String,
    val description: String?,
    val priceCents: Int,
    val category: String,
    val lat: Double,
    val lon: Double,
    val status: String,
    val createdAtEpochMillis: Long
)

@Serializable
data class CreateListingRequest(
    val sellerId: String,
    val title: String,
    val description: String? = null,
    val priceCents: Int,
    val category: String,
    val lat: Double,
    val lon: Double
)

/**
 * Straight Exposed DAO-less repository (functions over rows, not the
 * ActiveRecord-style DAO API) -- keeps the SQL visible and testable rather
 * than hidden behind entity magic.
 */
class ListingRepository {

    fun create(req: CreateListingRequest): Listing {
        val id = UUID.randomUUID().toString()
        val now = Instant.now()
        transaction {
            ListingTable.insert {
                it[ListingTable.id] = id
                it[sellerId] = req.sellerId
                it[title] = req.title
                it[description] = req.description
                it[priceCents] = req.priceCents
                it[category] = req.category
                it[lat] = req.lat
                it[lon] = req.lon
                it[status] = "ACTIVE"
                it[createdAt] = now
            }
        }
        return Listing(
            id = id,
            sellerId = req.sellerId,
            title = req.title,
            description = req.description,
            priceCents = req.priceCents,
            category = req.category,
            lat = req.lat,
            lon = req.lon,
            status = "ACTIVE",
            createdAtEpochMillis = now.toEpochMilli()
        )
    }

    fun findById(id: String): Listing? = transaction {
        ListingTable.selectAll().where { ListingTable.id eq id }
            .map { it.toListing() }
            .singleOrNull()
    }

    fun listRecent(limit: Int = 50): List<Listing> = transaction {
        ListingTable.selectAll()
            .orderBy(ListingTable.createdAt, SortOrder.DESC)
            .limit(limit)
            .map { it.toListing() }
    }

    fun markSold(id: String): Boolean = transaction {
        val updated = ListingTable.update({ ListingTable.id eq id }) {
            it[status] = "SOLD"
        }
        updated > 0
    }

    private fun ResultRow.toListing() = Listing(
        id = this[ListingTable.id],
        sellerId = this[ListingTable.sellerId],
        title = this[ListingTable.title],
        description = this[ListingTable.description],
        priceCents = this[ListingTable.priceCents],
        category = this[ListingTable.category],
        lat = this[ListingTable.lat],
        lon = this[ListingTable.lon],
        status = this[ListingTable.status],
        createdAtEpochMillis = this[ListingTable.createdAt].toEpochMilli()
    )
}
