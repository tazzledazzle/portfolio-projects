package com.marketplace.messaging

import io.lettuce.core.RedisClient
import io.lettuce.core.pubsub.RedisPubSubAdapter
import io.lettuce.core.pubsub.api.sync.RedisPubSubCommands
import kotlinx.coroutines.GlobalScope
import kotlinx.coroutines.launch
import kotlinx.serialization.decodeFromString
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json
import org.slf4j.LoggerFactory
import java.util.UUID
import java.util.concurrent.ConcurrentHashMap

/**
 * The concrete answer to "how does any pod find where a user's WebSocket
 * lives" (flagged as the hard part of messaging in the earlier design
 * review). Two Redis structures do the work:
 *
 *  1. A hash `presence` mapping userId -> podId, so any pod can look up
 *     which pod currently holds a user's socket.
 *  2. A pub/sub channel per pod (`pod:<podId>`); to deliver a message to a
 *     user connected elsewhere, this pod publishes to that user's pod's
 *     channel, and the owning pod fans it out to the local socket.
 *
 * This pod's own id is a random UUID generated at startup -- in the k8s
 * manifests, this doubles as a stand-in for the pod name.
 */
class ConnectionRegistry(redisUrl: String) {
    private val logger = LoggerFactory.getLogger(javaClass)
    private val json = Json { ignoreUnknownKeys = true }

    val podId: String = UUID.randomUUID().toString().take(8)

    private val redisClient = RedisClient.create(redisUrl)
    private val commands = redisClient.connect().sync()

    private val pubSubConnection = redisClient.connectPubSub()
    private val pubSubCommands: RedisPubSubCommands<String, String> = pubSubConnection.sync()

    // userId -> local WebSocket send function, only for sockets held by *this* pod.
    private val localSockets = ConcurrentHashMap<String, suspend (ChatMessage) -> Unit>()

    fun registerLocal(userId: String, sender: suspend (ChatMessage) -> Unit) {
        localSockets[userId] = sender
        commands.hset("presence", userId, podId)
        logger.info("User $userId connected to pod $podId")
    }

    fun unregisterLocal(userId: String) {
        localSockets.remove(userId)
        // Only clear presence if this pod still owns it -- avoids a race
        // where the user reconnected to a different pod microseconds ago.
        if (commands.hget("presence", userId) == podId) {
            commands.hdel("presence", userId)
        }
    }

    /**
     * Route a message to its recipient, wherever their socket lives.
     * Falls back to nothing (no-op) if the recipient is offline entirely --
     * the caller (ChatWebSocket) is still responsible for the Postgres
     * write, so the message isn't lost, just not delivered live.
     */
    suspend fun deliver(recipientId: String, message: ChatMessage) {
        val localSender = localSockets[recipientId]
        if (localSender != null) {
            localSender(message)
            return
        }
        val ownerPod = commands.hget("presence", recipientId) ?: return // offline
        val envelope = json.encodeToString(RoutedMessage(recipientId, message))
        commands.publish("pod:$ownerPod", envelope)
    }

    /**
     * Subscribes this pod to its own Redis pub/sub channel so messages
     * routed here from other pods (via [deliver]) reach locally-connected
     * sockets. Must be called once at startup.
     */
    fun startListeningForRoutedMessages() {
        pubSubConnection.addListener(object : RedisPubSubAdapter<String, String>() {
            override fun message(channel: String, message: String) {
                val routed = json.decodeFromString<RoutedMessage>(message)
                val sender = localSockets[routed.recipientId] ?: return
                // Pub/sub callback is synchronous; hand off is intentionally
                // fire-and-forget here to avoid blocking Lettuce's event loop.
                @OptIn(kotlinx.coroutines.DelicateCoroutinesApi::class)
                GlobalScope.launch { sender(routed.message) }
            }
        })
        pubSubCommands.subscribe("pod:$podId")
        logger.info("Pod $podId listening on channel pod:$podId")
    }

    fun close() {
        pubSubConnection.close()
        redisClient.shutdown()
    }
}

@kotlinx.serialization.Serializable
private data class RoutedMessage(val recipientId: String, val message: ChatMessage)
