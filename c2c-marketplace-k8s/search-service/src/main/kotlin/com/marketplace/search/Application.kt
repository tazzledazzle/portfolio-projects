package com.marketplace.search

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
import kotlinx.coroutines.GlobalScope
import kotlinx.coroutines.launch
import kotlinx.coroutines.runBlocking

fun main() {
    val openSearchUrl = System.getenv("OPENSEARCH_URL") ?: "http://localhost:9200"
    val kafkaBootstrap = System.getenv("KAFKA_BOOTSTRAP_SERVERS") ?: "localhost:9092"
    val port = System.getenv("PORT")?.toIntOrNull() ?: 8082

    val client = OpenSearchClient(openSearchUrl)
    runBlocking { client.ensureIndex() }

    val indexer = ListingIndexer(kafkaBootstrap, client)
    // Runs for the lifetime of the process alongside the HTTP server --
    // this is the "search-service is both an HTTP API and a Kafka consumer"
    // shape called out in the design review's container diagram.
    @OptIn(kotlinx.coroutines.DelicateCoroutinesApi::class)
    GlobalScope.launch { indexer.run() }

    embeddedServer(Netty, port = port, module = { module(client) }).start(wait = true)
}

fun Application.module(client: OpenSearchClient) {
    install(ContentNegotiation) { json() }
    install(CallLogging)
    routing {
        get("/healthz") { call.respond(mapOf("status" to "ok")) }
        searchRoutes(client)
    }
}
