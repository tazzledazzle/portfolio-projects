package com.skidroad.notifeed.inbox

import com.skidroad.notifeed.model.Notification
import com.skidroad.notifeed.model.toJson
import com.skidroad.notifeed.model.toNotification
import io.lettuce.core.api.StatefulRedisConnection
import org.slf4j.LoggerFactory

/**
 * InboxService
 * ────────────
 * Manages per-user notification inboxes stored as Redis Lists.
 *
 * Redis key pattern:  inbox:{userId}
 * Redis data type:    List
 *
 * Why a List?
 *  - LPUSH pushes new notifications to the HEAD (newest first without sorting).
 *  - LRANGE delivers a page of notifications with O(N) on the page size, not total length.
 *  - LTRIM prunes the list to a maximum depth in a single atomic command.
 *  - EXPIRE keeps inactive inboxes from accumulating forever in Redis memory.
 *
 * Flow:
 *   Subscriber receives Pub/Sub message
 *     → InboxService.push(notification)
 *         → LPUSH inbox:{userId} <json>
 *         → LTRIM inbox:{userId} 0 (MAX_INBOX_SIZE - 1)   // keep newest N
 *         → EXPIRE inbox:{userId} TTL_SECONDS              // refresh TTL on each write
 *
 * Reading:
 *   InboxService.getPage(userId, page, pageSize)
 *     → LRANGE inbox:{userId} start end
 *
 * @param conn         A general-purpose Lettuce command connection (thread-safe, shared).
 * @param maxDepth     Maximum notifications kept per inbox before LTRIM drops oldest.
 * @param ttlSeconds   How long an inbox lives without a new notification (default: 7 days).
 */
class InboxService(
    private val conn: StatefulRedisConnection<String, String>,
    private val maxDepth: Int = MAX_DEPTH,
    private val ttlSeconds: Long = TTL_SECONDS
) {
    private val log = LoggerFactory.getLogger(InboxService::class.java)
    private val commands = conn.sync()

    // ──────────────────────────────────────────
    // Write path
    // ──────────────────────────────────────────

    /**
     * Push a notification into [userId]'s inbox.
     *
     * LPUSH  — prepend JSON to the list (O(1))
     * LTRIM  — drop anything beyond [maxDepth] from the tail (O(N) on dropped elements)
     * EXPIRE — refresh the inbox TTL so inactive inboxes are eventually reclaimed
     */
    fun push(notification: Notification) {
        val key = inboxKey(notification.userId)
        val payload = notification.toJson()

        commands.lpush(key, payload)
        commands.ltrim(key, 0, (maxDepth - 1).toLong())
        commands.expire(key, ttlSeconds)

        log.debug(
            "LPUSH+LTRIM+EXPIRE → key='$key'  type=${notification.type}  " +
                    "depth=${commands.llen(key)}/$maxDepth"
        )
    }

    // ──────────────────────────────────────────
    // Read path
    // ──────────────────────────────────────────

    /**
     * Return one page of notifications from [userId]'s inbox, newest first.
     *
     * LRANGE key start stop  — O(N) on page size; the list is already sorted newest-first
     * because LPUSH prepends.
     *
     * @param userId   The inbox owner.
     * @param page     0-indexed page number.
     * @param pageSize Number of items per page (default 20).
     * @return         List of [Notification] objects, or empty list if inbox is absent.
     */
    fun getPage(userId: String, page: Int = 0, pageSize: Int = PAGE_SIZE): List<Notification> {
        val key = inboxKey(userId)
        val start = (page * pageSize).toLong()
        val stop = start + pageSize - 1

        return commands.lrange(key, start, stop)
            .mapNotNull { json ->
                runCatching { json.toNotification() }
                    .onFailure { log.warn("Failed to deserialise notification JSON: $json", it) }
                    .getOrNull()
            }
    }

    /**
     * Return the total number of notifications in [userId]'s inbox.
     * LLEN key — O(1)
     */
    fun size(userId: String): Long = commands.llen(inboxKey(userId))

    /**
     * Return the remaining TTL on the inbox key in seconds.
     * TTL key — O(1); returns -2 if key does not exist, -1 if no expiry set.
     */
    fun ttl(userId: String): Long = commands.ttl(inboxKey(userId))

    /**
     * Delete the inbox entirely (e.g. user logs out / clears notifications).
     * DEL key — O(N) on list length.
     */
    fun clear(userId: String): Long {
        val key = inboxKey(userId)
        val deleted = commands.del(key)
        log.info("DEL → key='$key'  result=$deleted")
        return deleted
    }

    // ──────────────────────────────────────────
    // Key naming
    // ──────────────────────────────────────────

    companion object {
        const val MAX_DEPTH = 100
        const val TTL_SECONDS = 7 * 24 * 60 * 60L  // 7 days
        const val PAGE_SIZE = 20

        fun inboxKey(userId: String) = "inbox:$userId"
    }
}