package com.marketplace.search

import com.marketplace.common.ListingCreatedEvent
import com.marketplace.common.Topics
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import kotlinx.serialization.decodeFromString
import kotlinx.serialization.json.Json
import org.apache.kafka.clients.consumer.ConsumerConfig
import org.apache.kafka.clients.consumer.KafkaConsumer
import org.apache.kafka.common.serialization.StringDeserializer
import org.slf4j.LoggerFactory
import java.time.Duration
import java.util.Properties

/**
 * Consumes listing.created off Kafka and writes into OpenSearch. This is
 * the whole reason search can stay eventually-consistent and independently
 * scaled from listings-service: if this consumer falls behind or restarts,
 * it resumes from its committed offset and catches up -- nothing is lost,
 * search is just briefly stale, which is an acceptable trade for a browse
 * feed (see TDD.md section 3 for why payments can't make the same trade).
 */
class ListingIndexer(
    bootstrapServers: String,
    private val openSearchClient: OpenSearchClient
) {
    private val logger = LoggerFactory.getLogger(javaClass)
    private val json = Json { ignoreUnknownKeys = true }

    private val consumer: KafkaConsumer<String, String> = KafkaConsumer(
        Properties().apply {
            put(ConsumerConfig.BOOTSTRAP_SERVERS_CONFIG, bootstrapServers)
            put(ConsumerConfig.GROUP_ID_CONFIG, "search-service")
            put(ConsumerConfig.KEY_DESERIALIZER_CLASS_CONFIG, StringDeserializer::class.java.name)
            put(ConsumerConfig.VALUE_DESERIALIZER_CLASS_CONFIG, StringDeserializer::class.java.name)
            put(ConsumerConfig.AUTO_OFFSET_RESET_CONFIG, "earliest")
            // Commit manually after a successful index write, not automatically
            // on a timer -- we want "indexed" and "offset committed" to be the
            // same fact, so a crash mid-batch replays instead of skipping.
            put(ConsumerConfig.ENABLE_AUTO_COMMIT_CONFIG, "false")
        }
    )

    suspend fun run() {
        consumer.subscribe(listOf(Topics.LISTING_CREATED))
        logger.info("ListingIndexer subscribed to ${Topics.LISTING_CREATED}")
        while (true) {
            val records = withContext(Dispatchers.IO) {
                consumer.poll(Duration.ofMillis(500))
            }
            for (record in records) {
                try {
                    val event = json.decodeFromString<ListingCreatedEvent>(record.value())
                    openSearchClient.indexListing(event)
                } catch (e: Exception) {
                    logger.error("Failed to index record at offset ${record.offset()}: ${e.message}", e)
                    // In a production version: dead-letter topic after N retries
                    // rather than blocking the partition forever.
                }
            }
            if (!records.isEmpty) {
                withContext(Dispatchers.IO) { consumer.commitSync() }
            }
        }
    }
}
