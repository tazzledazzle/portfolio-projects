package com.marketplace.synth

import kotlinx.coroutines.delay
import kotlin.random.Random

object Harness {
    private val tokenPattern = Regex("""#([\d-]+)""")

    suspend fun run(
        profile: Profile,
        client: MarketplaceClient,
        failFast: Boolean = false
    ): Summary {
        val random = Random(profile.seed)
        val gen = Generators(random)
        val errors = mutableListOf<String>()
        val createdListings = mutableListOf<CreatedListing>()

        var created = 0
        var indexed = 0
        var orders = 0
        var released = 0
        var refunded = 0

        fun summary() = Summary(
            profile = profile.name,
            created = created,
            indexed = indexed,
            orders = orders,
            released = released,
            refunded = refunded,
            chatOk = false,
            errors = errors.toList()
        )

        fun recordError(message: String): Boolean {
            errors.add(message)
            return failFast
        }

        for (i in 0 until profile.listings) {
            val sellerId = gen.sellerId(i)
            val title = gen.listingTitle(i, profile.categories)
            val priceCents = gen.priceCents()
            val category = gen.category(profile.categories)
            try {
                val id = client.createListing(
                    sellerId = sellerId,
                    title = title,
                    description = null,
                    priceCents = priceCents,
                    category = category,
                    lat = profile.geo.lat,
                    lon = profile.geo.lon
                )
                created++
                createdListings.add(CreatedListing(id, title, sellerId, priceCents))
            } catch (e: Exception) {
                if (recordError("createListing[$i]: ${e.message}")) return summary()
            }
        }

        for (listing in createdListings) {
            val token = distinctiveToken(listing.title)
            var found = false
            try {
                for (attempt in 0 until profile.searchRetries) {
                    val hits = client.search(token, profile.geo.lat, profile.geo.lon)
                    found = hits.any { hit ->
                        hit.title.contains(token) ||
                            hit.title == listing.title ||
                            hit.listingId == listing.id
                    }
                    if (found) break
                    if (attempt < profile.searchRetries - 1) {
                        delay(profile.searchRetryMs)
                    }
                }
            } catch (e: Exception) {
                if (recordError("search[${listing.id}]: ${e.message}")) return summary()
                continue
            }
            if (found) {
                indexed++
            } else if (recordError(
                    "listing ${listing.id} not indexed after ${profile.searchRetries} retries (q=$token)"
                )
            ) {
                return summary()
            }
        }

        val orderCount = minOf(profile.orders, createdListings.size)
        for (i in 0 until orderCount) {
            val listing = createdListings[i]
            val buyerId = gen.buyerId(i)
            try {
                val orderId = client.createOrder(
                    listingId = listing.id,
                    buyerId = buyerId,
                    sellerId = listing.sellerId,
                    amountCents = listing.priceCents
                )
                orders++
                val confirm = random.nextDouble() < profile.confirmRatio
                val status = if (confirm) {
                    client.confirmDelivery(orderId)
                } else {
                    client.dispute(orderId)
                }
                when (status) {
                    "RELEASED" -> released++
                    "REFUNDED" -> refunded++
                    else -> {
                        if (recordError("order $orderId unexpected status: $status")) return summary()
                    }
                }
            } catch (e: Exception) {
                if (recordError("order[$i listing=${listing.id}]: ${e.message}")) return summary()
            }
        }

        return summary()
    }

    internal fun distinctiveToken(title: String): String {
        val match = tokenPattern.find(title)
        return match?.value ?: title
    }

    private data class CreatedListing(
        val id: String,
        val title: String,
        val sellerId: String,
        val priceCents: Int
    )
}
