package com.marketplace.messaging

import com.zaxxer.hikari.HikariConfig
import com.zaxxer.hikari.HikariDataSource
import io.ktor.serialization.kotlinx.json.json
import io.ktor.server.application.Application
import io.ktor.server.application.install
import io.ktor.server.engine.embeddedServer
import io.ktor.server.netty.Netty
import io.ktor.server.plugins.callloging.CallLogging
import io.ktor.server.plugins.contentnegotiation.ContentNegotiation
import io.ktor.server.response.respond
import io.ktor.server.routing.get
import io.ktor.server.routing.routing
import io.ktor.server.websocket.WebSockets
import io.ktor.server.websocket.pingPeriod
import io.ktor.server.websocket.timeout
import org.jetbrains.exposed.sql.Database
import org.jetbrains.exposed.sql.SchemaUtils
import org.jetbrains.exposed.sql.transactions.transaction
import java.time.Duration

fun main() {
    val dbUrl = System.getenv("DB_URL") ?: "jdbc:postgresql://localhost:5432/marketplace"
    val dbUser = System.getenv("DB_USER") ?: "marketplace"
    val dbPassword = System.getenv("DB_PASSWORD") ?: "marketplace"
    val redisUrl = System.getenv("REDIS_URL") ?: "redis://localhost:6379"
    val port = System.getenv("PORT")?.toIntOrNull() ?: 8083

    val dataSource = HikariDataSource(HikariConfig().apply {
        jdbcUrl = dbUrl
        username = dbUser
        password = dbPassword
        maximumPoolSize = 10
    })
    Database.connect(dataSource)
    transaction { SchemaUtils.createMissingTablesAndColumns(MessageTable) }

    val repository = MessageRepository()
    val registry = ConnectionRegistry(redisUrl)
    registry.startListeningForRoutedMessages()

    embeddedServer(Netty, port = port, module = { module(repository, registry) }).start(wait = true)
}

fun Application.module(repository: MessageRepository, registry: ConnectionRegistry) {
    install(ContentNegotiation) { json() }
    install(CallLogging)
    install(WebSockets) {
        pingPeriod = Duration.ofSeconds(15)
        timeout = Duration.ofSeconds(30)
    }
    routing {
        get("/healthz") { call.respond(mapOf("status" to "ok", "podId" to registry.podId)) }
        get("/conversations/{id}/history") {
            val id = call.parameters["id"]!!
            call.respond(repository.history(id))
        }
        chatWebSocket(registry, repository)
    }
}
