package com.marketplace.messaging

import io.ktor.server.application.call
import io.ktor.server.routing.Route
import io.ktor.server.websocket.receiveDeserialized
import io.ktor.server.websocket.sendSerialized
import io.ktor.server.websocket.webSocket
import io.micrometer.core.instrument.MeterRegistry
import io.micrometer.core.instrument.Tag
import kotlinx.coroutines.channels.ClosedReceiveChannelException
import kotlinx.serialization.Serializable
import org.slf4j.LoggerFactory

@Serializable
data class OutgoingChatMessage(val conversationId: String, val body: String)

/**
 * One WebSocket connection per (userId). A real client would authenticate
 * via a token in the connection handshake; this mock takes userId as a
 * path parameter to keep the demo simple -- see TDD.md "known
 * simplifications" for what a real auth layer would add here.
 */
fun Route.chatWebSocket(
    registry: ConnectionRegistry,
    repository: MessageRepository,
    meters: MeterRegistry,
) {
    val logger = LoggerFactory.getLogger("ChatWebSocket")
    val connectCounter = { result: String ->
        meters.counter("messaging_ws_connect", listOf(Tag.of("result", result)))
    }
    val publishCounter = { result: String ->
        meters.counter("messaging_ws_publish", listOf(Tag.of("result", result)))
    }

    webSocket("/ws/{userId}") {
        val userId = call.parameters["userId"]!!

        try {
            registry.registerLocal(userId) { message ->
                sendSerialized(message)
            }
            connectCounter("success").increment()
        } catch (e: Exception) {
            connectCounter("failure").increment()
            throw e
        }

        try {
            while (true) {
                val incoming = receiveDeserialized<OutgoingChatMessage>()
                try {
                    val saved = repository.save(
                        conversationId = incoming.conversationId,
                        senderId = userId,
                        body = incoming.body
                    )
                    // Echo back to sender for optimistic-UI confirmation, then
                    // route to whichever pod (if any) holds the recipient's
                    // socket. This mock doesn't model conversation membership,
                    // so it derives "the other participant" from the
                    // conversationId being `<userA>:<userB>` -- see README.
                    sendSerialized(saved)
                    val recipientId = otherParticipant(incoming.conversationId, userId)
                    if (recipientId != null) {
                        registry.deliver(recipientId, saved)
                    }
                    publishCounter("success").increment()
                } catch (e: Exception) {
                    publishCounter("failure").increment()
                    throw e
                }
            }
        } catch (e: ClosedReceiveChannelException) {
            logger.info("User $userId disconnected")
        } finally {
            registry.unregisterLocal(userId)
        }
    }
}

private fun otherParticipant(conversationId: String, self: String): String? {
    val parts = conversationId.split(":")
    if (parts.size != 2) return null
    return parts.firstOrNull { it != self }
}
