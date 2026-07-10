package com.marketplace.search

import io.ktor.server.application.call
import io.ktor.server.response.respond
import io.ktor.server.routing.Route
import io.ktor.server.routing.get

fun Route.searchRoutes(client: OpenSearchClient) {
    get("/search") {
        val q = call.request.queryParameters["q"] ?: ""
        val lat = call.request.queryParameters["lat"]?.toDoubleOrNull()
        val lon = call.request.queryParameters["lon"]?.toDoubleOrNull()
        val radiusKm = call.request.queryParameters["radiusKm"]?.toIntOrNull() ?: 25

        val results = client.search(q, lat, lon, radiusKm)
        call.respond(results)
    }
}
