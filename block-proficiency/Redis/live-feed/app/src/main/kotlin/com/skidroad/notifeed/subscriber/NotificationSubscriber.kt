package com.skidroad.notifeed.subscriber

import com.skidroad.notifeed.inbox.InboxService
import com.skidroad.notifeed.model.toNotification
import com.skidroad.notifeed.publisher.NotificationPublisher
import io.lettuce.core.pubsub.RedisPubSubAdapter
import io.lettuce.core.pubsub.StatefulRedisPubSubConnection
import org.slf4j.LoggerFactory
import java.util.concurrent.atomic.AtomicLong

/**
 * NotificationSubscriber
 * ──────────────────────
 * Listens on the "notifications:events" Pub/Sub channel and routes each message
 * into the target user's List inbox via [InboxService].
 *
 * Lettuce Pub/Sub model:
 *   1. Attach a RedisPubSubAdapter listener to the connection.
 *   2. Call async().subscribe(channel) — the connection enters SUBSCRIBE mode.
 *   3. The listener's message() callback fires on every incoming PUBLISH.
 *
 * Threading:
 *   Lettuce delivers Pub/Sub messages on its internal I/O thread pool. The callback
 *   must not block that thread for long. Here we delegate synchronous Redis writes
 *   (LPUSH/LTRIM/EXPIRE) to InboxService which uses a separate command connection,
 *   so the I/O thread is released quickly.
 *
 * @param pubSubConn  Dedicated Lettuce Pub/Sub connection for SUBSCRIBE mode.
 * @param inboxService  Writes incoming notifications to per-user List inboxes.
 * @param channel     Channel to subscribe on. Must match [NotificationPublisher.CHANNEL].
 */
class NotificationSubscriber(
    private val pubSubConn: StatefulRedisPubSubConnection<String, String>,
    private val inboxService: InboxService,
    private val channel: String = NotificationPublisher.CHANNEL
) {
    private val log = LoggerFactory.getLogger(NotificationSubscriber::class.java)
    private val receivedCount = AtomicLong(0)

    /**
     * Register the listener and issue the SUBSCRIBE command.
     * Call once at startup; the subscription persists until [close] is called.
     */
    fun start() {
        pubSubConn.addListener(object : RedisPubSubAdapter<String, String>() {

            /**
             * Called by Lettuce for every PUBLISH received on [channel].
             *
             * @param channel  The channel name (useful when subscribing to multiple).
             * @param message  The raw JSON payload published by [NotificationPublisher].
             */
            override fun message(channel: String, message: String) {
                val count = receivedCount.incrementAndGet()
                log.debug("Pub/Sub message #$count on channel='$channel'")

                val notification = runCatching { message.toNotification() }
                    .onFailure { log.error("Failed to deserialise Pub/Sub message: $message", it) }
                    .getOrNull() ?: return

                log.info(
                    "→ Routing notification  id=${notification.id}  " +
                            "type=${notification.type}  userId=${notification.userId}"
                )

                // Write to the user's List inbox
                inboxService.push(notification)
            }

            /** Called when the SUBSCRIBE acknowledgement arrives from Redis. */
            override fun subscribed(channel: String, count: Long) {
                log.info("SUBSCRIBE confirmed  channel='$channel'  activeSubscriptions=$count")
            }

            /** Called when UNSUBSCRIBE is acknowledged. */
            override fun unsubscribed(channel: String, count: Long) {
                log.info("UNSUBSCRIBE confirmed  channel='$channel'  activeSubscriptions=$count")
            }
        })

        // Issue SUBSCRIBE asynchronously — returns immediately; messages arrive via listener
        pubSubConn.async().subscribe(channel)
        log.info("Subscriber started — listening on channel='$channel'")
    }

    /** Stop listening and release the Pub/Sub connection. */
    fun close() {
        pubSubConn.async().unsubscribe(channel)
        pubSubConn.close()
        log.info("Subscriber stopped  totalReceived=${receivedCount.get()}")
    }

    /** Diagnostic: total messages received since startup. */
    fun receivedCount(): Long = receivedCount.get()
}