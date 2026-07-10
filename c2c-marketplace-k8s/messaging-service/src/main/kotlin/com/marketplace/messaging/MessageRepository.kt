package com.marketplace.messaging

import kotlinx.serialization.Serializable
import org.jetbrains.exposed.sql.ResultRow
import org.jetbrains.exposed.sql.SortOrder
import org.jetbrains.exposed.sql.Table
import org.jetbrains.exposed.sql.insert
import org.jetbrains.exposed.sql.javatime.timestamp
import org.jetbrains.exposed.sql.selectAll
import org.jetbrains.exposed.sql.transactions.transaction
import java.time.Instant
import java.util.UUID

object MessageTable : Table("messages") {
    val id = varchar("id", 36)
    val conversationId = varchar("conversation_id", 36)
    val senderId = varchar("sender_id", 36)
    val body = text("body")
    val sentAt = timestamp("sent_at")

    override val primaryKey = PrimaryKey(id)
    init {
        index(isUnique = false, conversationId, sentAt)
    }
}

@Serializable
data class ChatMessage(
    val id: String,
    val conversationId: String,
    val senderId: String,
    val body: String,
    val sentAtEpochMillis: Long
)

/**
 * Postgres is a deliberate (if slightly unconventional) choice here over
 * Cassandra/DynamoDB, which is what the design review recommended for
 * production scale -- see TDD.md. For a single-node local mock, Postgres
 * is one less moving part; the access pattern (all messages for one
 * conversation, ordered by time) is identical either way, so swapping the
 * storage engine later wouldn't change this repository's interface.
 */
class MessageRepository {

    fun save(conversationId: String, senderId: String, body: String): ChatMessage {
        val id = UUID.randomUUID().toString()
        val now = Instant.now()
        transaction {
            MessageTable.insert {
                it[MessageTable.id] = id
                it[MessageTable.conversationId] = conversationId
                it[MessageTable.senderId] = senderId
                it[MessageTable.body] = body
                it[sentAt] = now
            }
        }
        return ChatMessage(id, conversationId, senderId, body, now.toEpochMilli())
    }

    fun history(conversationId: String, limit: Int = 100): List<ChatMessage> = transaction {
        MessageTable.selectAll().where { MessageTable.conversationId eq conversationId }
            .orderBy(MessageTable.sentAt, SortOrder.ASC)
            .limit(limit)
            .map { it.toChatMessage() }
    }

    private fun ResultRow.toChatMessage() = ChatMessage(
        id = this[MessageTable.id],
        conversationId = this[MessageTable.conversationId],
        senderId = this[MessageTable.senderId],
        body = this[MessageTable.body],
        sentAtEpochMillis = this[MessageTable.sentAt].toEpochMilli()
    )
}
