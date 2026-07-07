package com.skidroad.notifeed

import com.skidroad.notifeed.api.NotificationService
import com.skidroad.notifeed.inbox.InboxService
import com.skidroad.notifeed.model.NotificationType
import com.skidroad.notifeed.publisher.NotificationPublisher
import com.skidroad.notifeed.subscriber.NotificationSubscriber
import org.slf4j.LoggerFactory

private val log = LoggerFactory.getLogger("Main")

fun main() {
    log.info("=== Live Notification Feed — startup ===")

    // ── Step 1: Open Redis connections ──────────────────────────────────────
    //
    // We need THREE Lettuce connections:
    //   A. publisherConn  — for PUBLISH  (StatefulRedisPubSubConnection)
    //   B. subscriberConn — for SUBSCRIBE (StatefulRedisPubSubConnection)
    //   C. commandConn    — for LPUSH / LRANGE / LTRIM / EXPIRE (StatefulRedisConnection)
    //
    // Why separate connections for publisher and subscriber?
    //   Redis mandates that a connection in SUBSCRIBE mode can ONLY issue Pub/Sub commands.
    //   Mixing PUBLISH and SUBSCRIBE on the same connection causes protocol errors.

    val publisherConn   = RedisConfig.pubSubConnection()   // A
    val subscriberConn  = RedisConfig.pubSubConnection()   // B
    val commandConn     = RedisConfig.commandConnection()  // C

    // ── Step 2: Build service layer ──────────────────────────────────────────
    val publisher    = NotificationPublisher(publisherConn)
    val inboxService = InboxService(commandConn)
    val subscriber   = NotificationSubscriber(subscriberConn, inboxService)
    val service      = NotificationService(publisher, inboxService)

    // ── Step 3: Start subscriber (SUBSCRIBE on the Pub/Sub channel) ──────────
    subscriber.start()

    // Brief pause — SUBSCRIBE is async; give Lettuce time for the ACK from Redis
    Thread.sleep(200)

    // ── Step 4: Demo scenario ────────────────────────────────────────────────
    log.info("\n--- Publishing notifications for two users ---")

    // Alice gets an order and a mention
    service.send(
        userId   = "alice",
        type     = NotificationType.ORDER_SHIPPED,
        title    = "Your order #1042 has shipped!",
        body     = "Estimated delivery: 2 business days.",
        metadata = mapOf("orderId" to "1042", "carrier" to "FedEx")
    )
    service.send(
        userId = "alice",
        type   = NotificationType.MENTION,
        title  = "Bob mentioned you in a comment",
        body   = "@alice great work on the Q2 report!"
    )

    // Bob gets a friend request and a system alert
    service.send(
        userId = "bob",
        type   = NotificationType.FRIEND_REQUEST,
        title  = "Alice sent you a friend request"
    )
    service.send(
        userId = "bob",
        type   = NotificationType.SYSTEM_ALERT,
        title  = "Scheduled maintenance tonight at 11 PM UTC",
        body   = "Expected downtime: 30 minutes."
    )

    // Wait for all Pub/Sub messages to be delivered to the subscriber and written to inboxes
    Thread.sleep(300)

    // ── Step 5: Read inboxes ─────────────────────────────────────────────────
    log.info("\n--- Reading inboxes ---")

    printInbox(service, "alice")
    printInbox(service, "bob")

    // ── Step 6: Demonstrate pagination ──────────────────────────────────────
    log.info("\n--- Sending 25 more notifications to alice to demonstrate pagination ---")
    repeat(25) { i ->
        service.send(
            userId = "alice",
            type   = NotificationType.COMMENT_ADDED,
            title  = "New comment on your post #${i + 1}"
        )
    }
    Thread.sleep(500)

    log.info("\nAlice inbox size: ${service.inboxSize("alice")}")
    log.info("Page 0 (newest 20):")
    service.inbox("alice", page = 0).forEachIndexed { i, n ->
        log.info("  [$i] ${n.type} — ${n.title}")
    }
    log.info("Page 1 (next 7):")
    service.inbox("alice", page = 1).forEachIndexed { i, n ->
        log.info("  [$i] ${n.type} — ${n.title}")
    }

    // ── Step 7: TTL inspection ───────────────────────────────────────────────
    log.info("\n--- TTL inspection ---")
    log.info("alice inbox TTL: ${service.inboxTtl("alice")} seconds (≈7 days)")
    log.info("bob   inbox TTL: ${service.inboxTtl("bob")} seconds")

    // ── Step 8: Clear an inbox ───────────────────────────────────────────────
    log.info("\n--- Clearing bob's inbox ---")
    service.clearInbox("bob")
    log.info("bob inbox size after clear: ${service.inboxSize("bob")}")
    log.info("bob inbox TTL  after clear: ${service.inboxTtl("bob")} (−2 = key gone)")

    // ── Step 9: Subscriber stats ─────────────────────────────────────────────
    log.info("\nTotal Pub/Sub messages received by subscriber: ${subscriber.receivedCount()}")

    // ── Shutdown ─────────────────────────────────────────────────────────────
    log.info("\n=== Shutting down ===")
    subscriber.close()
    publisher.close()
    commandConn.close()
    RedisConfig.shutdown()

    log.info("Done.")
}

private fun printInbox(service: NotificationService, userId: String) {
    val notifications = service.inbox(userId)
    val size = service.inboxSize(userId)
    val ttl  = service.inboxTtl(userId)
    log.info("\n  Inbox for '$userId'  (size=$size  ttl=${ttl}s)")
    if (notifications.isEmpty()) {
        log.info("    (empty)")
    } else {
        notifications.forEachIndexed { i, n ->
            log.info("    [$i] ${n.type.name.padEnd(20)} | ${n.title}")
        }
    }
}