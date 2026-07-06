package com.skidroad.notifeed.api

import com.skidroad.notifeed.inbox.InboxService
import com.skidroad.notifeed.model.Notification
import com.skidroad.notifeed.model.NotificationType
import com.skidroad.notifeed.publisher.NotificationPublisher
import org.slf4j.LoggerFactory
import java.util.UUID

/**
 * NotificationService
 * ───────────────────
 * High-level facade that wires Publisher and InboxService together.
 * This is the entry point for application code — it hides Redis keys and
 * channel names behind a clean domain API.
 *
 * Usage:
 *   service.send(userId = "alice", type = ORDER_SHIPPED, title = "Your order is on the way")
 *   service.inbox(userId = "alice", page = 0)
 *   service.inboxSize(userId = "alice")
 *   service.inboxTtl(userId = "alice")
 *   service.clearInbox(userId = "alice")
 */
class NotificationService(
    private val publisher: NotificationPublisher,
    private val inboxService: InboxService
) {
    private val log = LoggerFactory.getLogger(NotificationService::class.java)

    /**
     * Create and publish a notification for [userId].
     * The subscriber picks it up and writes it into the user's List inbox.
     */
    fun send(
        userId: String,
        type: NotificationType,
        title: String,
        body: String = "",
        metadata: Map<String, String> = emptyMap()
    ): Notification {
        val notification = Notification(
            id = UUID.randomUUID().toString(),
            type = type,
            userId = userId,
            title = title,
            body = body,
            metadata = metadata
        )
        val receiverCount = publisher.publish(notification)
        log.info("Sent notification  id=${notification.id}  receivers=$receiverCount")
        return notification
    }

    /**
     * Retrieve a page from [userId]'s inbox.
     * Newest notifications appear first (LPUSH ordering).
     */
    fun inbox(userId: String, page: Int = 0, pageSize: Int = InboxService.PAGE_SIZE): List<Notification> =
        inboxService.getPage(userId, page, pageSize)

    /** Number of notifications in [userId]'s inbox. */
    fun inboxSize(userId: String): Long = inboxService.size(userId)

    /**
     * Remaining TTL on [userId]'s inbox key in seconds.
     * -2 means the key does not exist; -1 means no expiry (shouldn't happen here).
     */
    fun inboxTtl(userId: String): Long = inboxService.ttl(userId)

    /** Delete all notifications for [userId]. */
    fun clearInbox(userId: String): Long = inboxService.clear(userId)
}