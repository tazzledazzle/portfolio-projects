package com.skidroad.notifeed.model

import java.time.Instant

/**
 * Notification — the unit of work flowing through Pub/Sub and into each user's List inbox.
 *
 * Redis data structures used:
 *  - Pub/Sub channel  : "notifications:events"  (fan-out to all active subscribers)
 *  - List inbox key   : "inbox:{userId}"         (per-user durable queue)
 *
 * The payload is JSON-serialised for transport. Redis Strings hold the JSON; the List
 * stores those strings as elements.
 */
data class Notification(
    val id: String,
    val type: NotificationType,
    val userId: String,          // target recipient
    val title: String,
    val body: String = "",
    val metadata: Map<String, String> = emptyMap(),
    val createdAt: Instant = Instant.now(),
    val read: Boolean = false
)

enum class NotificationType {
    ORDER_SHIPPED,
    COMMENT_ADDED,
    MENTION,
    SYSTEM_ALERT,
    FRIEND_REQUEST
}