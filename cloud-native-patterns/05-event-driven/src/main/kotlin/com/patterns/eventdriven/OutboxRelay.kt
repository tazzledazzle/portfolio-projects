package com.patterns.eventdriven

import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.cancel
import kotlinx.coroutines.delay
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch
import org.slf4j.LoggerFactory
import java.util.concurrent.atomic.AtomicBoolean
import java.util.concurrent.atomic.AtomicInteger

/**
 * Simulated CDC (Change Data Capture) relay.
 *
 * In production: Debezium monitors the PostgreSQL WAL (Write-Ahead Log) and
 * publishes change events to Kafka automatically — no polling needed. This
 * in-memory relay simulates the same semantics with a polling loop.
 *
 * The key guarantee this relay provides:
 *   If [publishToKafka] fails for any reason (Kafka unavailable, network error),
 *   the outbox event remains unpublished (published=false) and will be retried
 *   on the next poll cycle. Events are never lost once written to the outbox.
 *
 * Delivery guarantee: at-least-once. The same event may be published more than
 * once if the relay crashes between publish and markPublished. Consumers must
 * be idempotent (deduplicate by event.id).
 */
class OutboxRelay(
    private val outboxRepository: OutboxRepository,
    private val pollIntervalMs: Long = 1_000L,
    private val simulateFailureEveryN: Int = 0,  // 0 = no simulated failures
) {
    private val log = LoggerFactory.getLogger(OutboxRelay::class.java)
    private val scope = CoroutineScope(Dispatchers.Default + SupervisorJob())
    private val publishCallCount = AtomicInteger(0)
    private val running = AtomicBoolean(false)
    private var job: Job? = null

    fun start() {
        running.set(true)
        job = scope.launch {
            log.info("[OutboxRelay] Starting — polling every {}ms", pollIntervalMs)
            while (isActive && running.get()) {
                poll()
                delay(pollIntervalMs)
            }
            log.info("[OutboxRelay] Stopped.")
        }
    }

    fun stop() {
        running.set(false)
        scope.cancel()
        log.info("[OutboxRelay] Stop requested.")
    }

    /** Single poll cycle — processes all currently unpublished events. */
    suspend fun poll() {
        val unpublished = outboxRepository.findUnpublished()
        if (unpublished.isEmpty()) return

        log.debug("[OutboxRelay] Found {} unpublished event(s)", unpublished.size)

        for (event in unpublished) {
            try {
                outboxRepository.incrementAttempts(event.id)
                publishToKafka(event)
                outboxRepository.markPublished(event.id)
                log.info(
                    "[OutboxRelay] Published event id={} type={} aggregateId={} topic=orders.{}",
                    event.id,
                    event.eventType,
                    event.aggregateId,
                    event.aggregateType.lowercase(),
                )
            } catch (e: Exception) {
                // Publish failed — leave the event unpublished for next poll cycle.
                // This is the core of the outbox guarantee: we never lose the event.
                log.warn(
                    "[OutboxRelay] Publish failed for event id={} type={} attempt={}: {}",
                    event.id,
                    event.eventType,
                    event.publishAttempts + 1,
                    e.message,
                )
            }
        }
    }

    /**
     * Simulates publishing an event to Kafka.
     *
     * In production: KafkaProducer.send(ProducerRecord(topicName, aggregateId, payload))
     * where aggregateId is the message key (guarantees per-entity partition ordering).
     *
     * The Kafka topic name is derived from the aggregate type — consistent with the
     * Debezium naming convention: <connector>.<schema>.<table>
     */
    private fun publishToKafka(event: OutboxEvent) {
        val callN = publishCallCount.incrementAndGet()

        // Simulate intermittent Kafka publish failures
        if (simulateFailureEveryN > 0 && callN % simulateFailureEveryN == 0) {
            throw RuntimeException("Simulated Kafka publish failure (call #$callN)")
        }

        val topicName = "orders.${event.aggregateType.lowercase()}.${event.eventType.lowercase()}"
        // Key = aggregateId ensures all events for the same Order go to the same partition
        val messageKey = event.aggregateId

        // Simulate network latency
        Thread.sleep(10)

        println(
            """
            |[Kafka] → Topic: $topicName
            |         Key:   $messageKey
            |         Value: ${event.payload}
            |         EventId: ${event.id}
            """.trimMargin()
        )
    }

    fun isRunning(): Boolean = running.get()
}
