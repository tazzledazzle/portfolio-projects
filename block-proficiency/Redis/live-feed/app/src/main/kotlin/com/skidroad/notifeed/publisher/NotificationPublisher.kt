package com.skidroad.notifeed.publisher

import com.skidroad.notifeed.model.Notification
import com.skidroad.notifeed.model.toJson
import io.lettuce.core.pubsub.StatefulRedisPubSubConnection
import org.slf4j.LoggerFactory

/**
 * NotificationPublisher
 * ─────────────────────
 * Publishes notification events onto a Redis Pub/Sub channel.
 *
 * Redis pattern:
 *   PUBLISH notifications:events <json>
 *
 * Every active subscriber listening on "notifications:events" receives the message
 * immediately. This is pure fan-out — Redis does not persist the message on the channel.
 * Durability is the subscriber's responsibility (it writes to the user's List inbox).
 *
 * Why a dedicated Pub/Sub connection?
 *   Redis requires that a connection in SUBSCRIBE mode can only issue Pub/Sub commands.
 *   The publisher needs to PUBLISH, not subscribe, so it uses a separate connection but
 *   still of type StatefulRedisPubSubConnection (which supports both roles).
 *
 * @param pubSubConn  A Lettuce Pub/Sub connection (one per publisher is fine).
 * @param channel     The channel name to publish on. Default: "notifications:events".
 */
class NotificationPublisher(
    private val pubSubConn: StatefulRedisPubSubConnection<String, String>,
    private val channel: String = CHANNEL
) {
    private val log = LoggerFactory.getLogger(NotificationPublisher::class.java)

    // Synchronous Pub/Sub commands — simple and predictable for publishing
    private val commands = pubSubConn.sync()

    /**
     * Publish a [Notification] to the channel.
     * Returns the number of subscribers that received the message.
     */
    fun publish(notification: Notification): Long {
        val payload = notification.toJson()
        val receiverCount = commands.publish(channel, payload)
        log.info(
            "PUBLISH → channel='$channel'  type=${notification.type}  " +
                    "userId=${notification.userId}  receivers=$receiverCount"
        )
        return receiverCount
    }

    fun close() {
        pubSubConn.close()
    }

    companion object {
        const val CHANNEL = "notifications:events"
    }
}