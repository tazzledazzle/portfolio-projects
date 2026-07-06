package com.skidroad.notifeed

import com.skidroad.notifeed.inbox.InboxService
import com.skidroad.notifeed.model.Notification
import com.skidroad.notifeed.model.NotificationType
import com.skidroad.notifeed.model.toJson
import io.lettuce.core.api.StatefulRedisConnection
import io.lettuce.core.api.sync.RedisCommands
import io.mockk.*
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import java.util.UUID
import kotlin.test.assertEquals

class InboxServiceTest {

    private val conn     = mockk<StatefulRedisConnection<String, String>>()
    private val commands = mockk<RedisCommands<String, String>>()

    private lateinit var inboxService: InboxService

    @BeforeEach
    fun setUp() {
        every { conn.sync() } returns commands
        inboxService = InboxService(conn, maxDepth = 5, ttlSeconds = 3600)
    }

    // ─── push() ──────────────────────────────────────────────────────────────

    @Test
    fun `push calls LPUSH, LTRIM, EXPIRE and LLEN in order`() {
        val notification = makeNotification("user-1")
        val key = InboxService.inboxKey("user-1")

        every { commands.lpush(key, any<String>()) } returns 1L
        every { commands.ltrim(key, 0, 4) } returns "OK"
        every { commands.expire(key, 3600L) } returns true
        every { commands.llen(key) } returns 1L

        inboxService.push(notification)

        verifyOrder {
            commands.lpush(key, any<String>())
            commands.ltrim(key, 0, 4)      // maxDepth=5, so stop=4
            commands.expire(key, 3600L)
            commands.llen(key)             // called by debug log
        }
    }

    @Test
    fun `push uses correct inbox key pattern`() {
        val notification = makeNotification("alice-99")
        val expectedKey = "inbox:alice-99"

        every { commands.lpush(expectedKey, any<String>()) } returns 1L
        every { commands.ltrim(expectedKey, 0, 4) } returns "OK"
        every { commands.expire(expectedKey, 3600L) } returns true
        every { commands.llen(expectedKey) } returns 1L

        inboxService.push(notification)

        verify { commands.lpush(expectedKey, any<String>()) }
    }

    // ─── getPage() ───────────────────────────────────────────────────────────

    @Test
    fun `getPage returns correct page from LRANGE`() {
        val userId = "bob"
        val notifications = (1..3).map { makeNotification(userId) }
        val jsonList = notifications.map { it.toJson() }
        val key = InboxService.inboxKey(userId)

        every { commands.lrange(key, 0, 19) } returns jsonList

        val result = inboxService.getPage(userId, page = 0, pageSize = 20)

        assertEquals(3, result.size)
        assertEquals(notifications[0].id, result[0].id)
    }

    @Test
    fun `getPage computes correct LRANGE start and stop for page 2`() {
        val userId = "carol"
        val key = InboxService.inboxKey(userId)

        every { commands.lrange(key, 40, 59) } returns emptyList()

        inboxService.getPage(userId, page = 2, pageSize = 20)

        verify { commands.lrange(key, 40, 59) }
    }

    @Test
    fun `getPage skips malformed JSON and returns valid items`() {
        val userId = "dave"
        val good = makeNotification(userId).toJson()
        val key = InboxService.inboxKey(userId)

        every { commands.lrange(key, 0, 19) } returns listOf(good, "not-valid-json", good)

        val result = inboxService.getPage(userId)

        assertEquals(2, result.size)
    }

    // ─── size() and ttl() ────────────────────────────────────────────────────

    @Test
    fun `size delegates to LLEN`() {
        every { commands.llen("inbox:eve") } returns 42L
        assertEquals(42L, inboxService.size("eve"))
    }

    @Test
    fun `ttl delegates to TTL command`() {
        every { commands.ttl("inbox:frank") } returns 86400L
        assertEquals(86400L, inboxService.ttl("frank"))
    }

    // ─── clear() ─────────────────────────────────────────────────────────────

    @Test
    fun `clear calls DEL on the inbox key`() {
        val key = "inbox:grace"
        every { commands.del(key) } returns 1L

        val result = inboxService.clear("grace")

        assertEquals(1L, result)
        verify { commands.del(key) }
    }

    // ─── helpers ─────────────────────────────────────────────────────────────

    private fun makeNotification(userId: String) = Notification(
        id     = UUID.randomUUID().toString(),
        type   = NotificationType.COMMENT_ADDED,
        userId = userId,
        title  = "Test notification"
    )
}