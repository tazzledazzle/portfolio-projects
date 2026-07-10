package com.marketplace.payments

import com.marketplace.common.OrderCompletedEvent
import com.marketplace.common.OrderCreatedEvent
import com.marketplace.common.Topics
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json
import org.apache.kafka.clients.producer.KafkaProducer
import org.apache.kafka.clients.producer.ProducerConfig
import org.apache.kafka.clients.producer.ProducerRecord
import org.apache.kafka.common.serialization.StringSerializer
import org.slf4j.LoggerFactory
import java.time.Instant
import java.util.Properties

/**
 * Published *after* the escrow transaction commits, never inside it -- see
 * OrderRepository. The order/hold rows in Postgres are the source of
 * truth; these events are a best-effort notification for anything that
 * wants to react (a notifications service, in a fuller build).
 */
class EventPublisher(bootstrapServers: String) {
    private val logger = LoggerFactory.getLogger(javaClass)
    private val json = Json { ignoreUnknownKeys = true }

    private val producer: KafkaProducer<String, String> = KafkaProducer(
        Properties().apply {
            put(ProducerConfig.BOOTSTRAP_SERVERS_CONFIG, bootstrapServers)
            put(ProducerConfig.KEY_SERIALIZER_CLASS_CONFIG, StringSerializer::class.java.name)
            put(ProducerConfig.VALUE_SERIALIZER_CLASS_CONFIG, StringSerializer::class.java.name)
            put(ProducerConfig.ACKS_CONFIG, "all")
        }
    )

    fun publishOrderCreated(order: Order) {
        val event = OrderCreatedEvent(
            orderId = order.id,
            listingId = order.listingId,
            buyerId = order.buyerId,
            sellerId = order.sellerId,
            amountCents = order.amountCents,
            createdAt = Instant.now().toEpochMilli()
        )
        send(Topics.ORDER_CREATED, order.id, json.encodeToString(event))
    }

    fun publishOrderCompleted(orderId: String) {
        val event = OrderCompletedEvent(orderId = orderId, completedAt = Instant.now().toEpochMilli())
        send(Topics.ORDER_COMPLETED, orderId, json.encodeToString(event))
    }

    private fun send(topic: String, key: String, payload: String) {
        producer.send(ProducerRecord(topic, key, payload)) { metadata, exception ->
            if (exception != null) {
                logger.error("Failed to publish to $topic for key=$key", exception)
            } else {
                logger.info("Published $topic key=$key offset=${metadata.offset()}")
            }
        }
    }

    fun close() = producer.close()
}
