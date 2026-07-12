package com.marketplace.payments

import com.marketplace.common.observability.Observability
import com.marketplace.common.observability.installObservability
import com.zaxxer.hikari.HikariConfig
import com.zaxxer.hikari.HikariDataSource
import io.ktor.http.HttpStatusCode
import io.ktor.serialization.kotlinx.json.json
import io.ktor.server.application.Application
import io.ktor.server.application.install
import io.ktor.server.engine.embeddedServer
import io.ktor.server.netty.Netty
import io.ktor.server.plugins.callloging.CallLogging
import io.ktor.server.plugins.contentnegotiation.ContentNegotiation
import io.ktor.server.plugins.statuspages.StatusPages
import io.ktor.server.application.call
import io.ktor.server.response.respond
import io.ktor.server.routing.get
import io.ktor.server.routing.routing
import io.micrometer.prometheusmetrics.PrometheusMeterRegistry
import io.opentelemetry.api.OpenTelemetry
import org.jetbrains.exposed.sql.Database
import org.jetbrains.exposed.sql.SchemaUtils
import org.jetbrains.exposed.sql.transactions.transaction

private const val SERVICE_NAME = "payments-service"

fun main() {
    val dbUrl = System.getenv("DB_URL") ?: "jdbc:postgresql://localhost:5432/marketplace"
    val dbUser = System.getenv("DB_USER") ?: "marketplace"
    val dbPassword = System.getenv("DB_PASSWORD") ?: "marketplace"
    val kafkaBootstrap = System.getenv("KAFKA_BOOTSTRAP_SERVERS") ?: "localhost:9092"
    val port = System.getenv("PORT")?.toIntOrNull() ?: 8084

    val registry = Observability.createPrometheusRegistry(SERVICE_NAME)
    val openTelemetry = Observability.initOpenTelemetry(SERVICE_NAME)

    val dataSource = HikariDataSource(HikariConfig().apply {
        jdbcUrl = dbUrl
        username = dbUser
        password = dbPassword
        maximumPoolSize = 10
    })
    Database.connect(dataSource)
    transaction {
        SchemaUtils.createMissingTablesAndColumns(OrderTable, EscrowHoldTable)
    }

    val repository = OrderRepository()
    val publisher = EventPublisher(kafkaBootstrap)

    embeddedServer(Netty, port = port, module = {
        module(repository, publisher, registry, openTelemetry)
    }).start(wait = true)
}

fun Application.module(
    repository: OrderRepository,
    publisher: EventPublisher,
    meterRegistry: PrometheusMeterRegistry = Observability.createPrometheusRegistry(SERVICE_NAME),
    openTelemetry: OpenTelemetry = Observability.initOpenTelemetry(SERVICE_NAME),
) {
    installObservability(SERVICE_NAME, meterRegistry, openTelemetry)
    install(ContentNegotiation) { json() }
    install(CallLogging)
    install(StatusPages) {
        exception<Throwable> { call, cause ->
            call.respond(HttpStatusCode.InternalServerError, ErrorResponse(cause.message ?: "internal error"))
        }
    }
    routing {
        get("/healthz") { call.respond(mapOf("status" to "ok")) }
        paymentsRoutes(repository, publisher, meterRegistry)
    }
}
