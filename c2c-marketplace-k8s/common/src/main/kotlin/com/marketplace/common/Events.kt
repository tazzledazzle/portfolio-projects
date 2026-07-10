package com.marketplace.common

import kotlinx.serialization.Serializable

/**
 * Shared event schemas published to Kafka/Redpanda.
 *
 * These live in one module so listings-service (producer) and search-service
 * (consumer) can't silently drift on field names or types. In a larger org
 * this would likely be a versioned schema registry (Avro/Protobuf) instead
 * of a shared Kotlin module across services owned by different teams — for
 * a handful of services owned by one person, a shared module is the simpler
 * choice and keeps compile-time safety.
 */
@Serializable
data class ListingCreatedEvent(
    val listingId: String,
    val sellerId: String,
    val title: String,
    val description: String?,
    val priceCents: Int,
    val category: String,
    val lat: Double,
    val lon: Double,
    val createdAt: Long
)

@Serializable
data class ListingUpdatedEvent(
    val listingId: String,
    val title: String,
    val description: String?,
    val priceCents: Int,
    val category: String,
    val status: String,
    val updatedAt: Long
)

@Serializable
data class OrderCreatedEvent(
    val orderId: String,
    val listingId: String,
    val buyerId: String,
    val sellerId: String,
    val amountCents: Int,
    val createdAt: Long
)

@Serializable
data class OrderCompletedEvent(
    val orderId: String,
    val completedAt: Long
)

object Topics {
    const val LISTING_CREATED = "listing.created"
    const val LISTING_UPDATED = "listing.updated"
    const val ORDER_CREATED = "order.created"
    const val ORDER_COMPLETED = "order.completed"
}
