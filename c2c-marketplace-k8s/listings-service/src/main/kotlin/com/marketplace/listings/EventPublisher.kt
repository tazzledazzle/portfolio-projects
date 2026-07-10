package com.marketplace.listings

import com.marketplace.common.ListingCreatedEvent
import com.marketplace.common.Topics
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json
import org.apache.kafka.clients.producer.KafkaProducer
import org.apache.kafka.clients.producer.ProducerConfig
import org.apache.kafka.clients.producer.ProducerRecord
import org.apache.kafka.common.serialization.StringSerializer
import org.slf4j.LoggerFactory
import java.util.Properties

/**
 * Publishes listing lifecycle events to Kafka/Redpanda so search-service
 * (and anything else that shows up later -- notifications, analytics) can
 * consume off the stream instead of listings-service having to know who's
 * listening. Keyed by listingId so ordering per-listing is preserved.
 */
class EventPublisher(bootstrapServers: String) {
    private val logger = LoggerFactory.getLogger(javaClass)
    private val json = Json { ignoreUnknownKeys = true }

    private val producer: KafkaProducer<String, String> = KafkaProducer(
        Properties().apply {
            put(ProducerConfig.BOOTSTRAP_SERVERS_CONFIG, bootstrapServers)
            put(ProducerConfig.KEY_SERIALIZER_CLASS_CONFIG, StringSerializer::class.java.name)
            put(ProducerConfig.VALUE_SERIALIZER_CLASS_CONFIG, StringSerializer::class.java.name)
            // acks=all: we'd rather block briefly than silently lose a
            // listing.created event -- search staying in sync depends on it.
            put(ProducerConfig.ACKS_CONFIG, "all")
        }
    )

    fun publishListingCreated(listing: Listing) {
        val event = ListingCreatedEvent(
            listingId = listing.id,
            sellerId = listing.sellerId,
            title = listing.title,
            description = listing.description,
            priceCents = listing.priceCents,
            category = listing.category,
            lat = listing.lat,
            lon = listing.lon,
            createdAt = listing.createdAtEpochMillis
        )
        val payload = json.encodeToString(event)
        producer.send(ProducerRecord(Topics.LISTING_CREATED, listing.id, payload)) { metadata, exception ->
            if (exception != null) {
                logger.error("Failed to publish listing.created for ${listing.id}", exception)
            } else {
                logger.info("Published listing.created id=${listing.id} partition=${metadata.partition()} offset=${metadata.offset()}")
            }
        }
    }

    fun close() = producer.close()
}
