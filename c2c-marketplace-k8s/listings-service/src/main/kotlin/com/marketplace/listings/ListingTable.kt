package com.marketplace.listings

import org.jetbrains.exposed.sql.Table
import org.jetbrains.exposed.sql.javatime.timestamp

object ListingTable : Table("listings") {
    val id = varchar("id", 36)
    val sellerId = varchar("seller_id", 36)
    val title = varchar("title", 200)
    val description = text("description").nullable()
    val priceCents = integer("price_cents")
    val category = varchar("category", 64)
    val lat = double("lat")
    val lon = double("lon")
    val status = varchar("status", 16).default("ACTIVE")
    val createdAt = timestamp("created_at")

    override val primaryKey = PrimaryKey(id)
}
