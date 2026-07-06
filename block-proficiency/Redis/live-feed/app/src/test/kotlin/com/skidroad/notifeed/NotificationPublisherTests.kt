package com.skidroad.notifeed

import com.skidroad.notifeed.model.Notification
import com.skidroad.notifeed.model.NotificationType
import com.skidroad.notifeed.publisher.NotificationPublisher
import io.lettuce.core.pubsub.StatefulRedisPubSubConnection
import io.lettuce.core.pubsub.api.sync.RedisPubSubCommands
import io.mockk.*
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import java.util.UUID
import kotlin.test.assertEquals

class NotificationPublisherTest {

    private val pubSubConn = mockk<StatefulRedisPubSubConnection<String, String>>()
    private val pubSubSync = mockk<RedisPubSubCommands<String, String>>()

    private lateinit var publisher: NotificationPublisher

    @BeforeEach
    fun setUp() {
        every { pubSubConn.sync() } returns pubSubSync
        publisher = NotificationPublisher(pubSubConn)
    }

    @Test
    fun `publish calls PUBLISH on the correct channel`() {
        every { pubSubSync.publish(NotificationPublisher.CHANNEL, any()) } returns 3L

        val notification = makeNotification("alice")
        val count = publisher.publish(notification)

        assertEquals(3L, count)
        verify { pubSubSync.publish(NotificationPublisher.CHANNEL, any()) }
    }

    @Test
    fun `publish returns 0 when no subscribers are active`() {
        every { pubSubSync.publish(any(), any()) } returns 0L

        val count = publisher.publish(makeNotification("bob"))
        assertEquals(0L, count)
    }

    @Test
    fun `publish serialises the notification as JSON containing the userId`() {
        val slot = slot<String>()
        every { pubSubSync.publish(any(), capture(slot)) } returns 1L

        publisher.publish(makeNotification("charlie"))

        assert(slot.captured.contains("charlie")) {
            "Expected published JSON to contain userId 'charlie', got: ${slot.captured}"
        }
    }

    private fun makeNotification(userId: String) = Notification(
        id     = UUID.randomUUID().toString(),
        type   = NotificationType.ORDER_SHIPPED,
        userId = userId,
        title  = "Test"
    )
}