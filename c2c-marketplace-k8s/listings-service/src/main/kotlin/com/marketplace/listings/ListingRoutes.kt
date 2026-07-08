package com.marketplace.listings

import io.ktor.http.HttpStatusCode
import io.ktor.server.application.call
import io.ktor.server.request.receive
import io.ktor.server.response.respond
import io.ktor.server.routing.Route
import io.ktor.server.routing.get
import io.ktor.server.routing.post
import kotlinx.serialization.Serializable

/**
 * Stub trust & safety check, standing in for OfferUp's real ML-driven fraud
 * and counterfeit detection. Real version: async, model-scored, escalates
 * to human review. This version: synchronous keyword blocklist, good enough
 * to demonstrate "listings service owns a T&S gate before publish" without
 * pretending to be a fraud model.
 */
private val BLOCKED_KEYWORDS = setOf("stolen", "counterfeit", "replica")

private fun trustAndSafetyCheck(req: CreateListingRequest): String? {
    val haystack = "${req.title} ${req.description.orEmpty()}".lowercase()
    val hit = BLOCKED_KEYWORDS.firstOrNull { haystack.contains(it) }
    return hit?.let { "listing blocked: contains disallowed term '$it'" }
}

@Serializable
data class ErrorResponse(val error: String)

fun Route.listingRoutes(repository: ListingRepository, publisher: EventPublisher) {

    post("/listings") {
        val req = call.receive<CreateListingRequest>()

        trustAndSafetyCheck(req)?.let { reason ->
            call.respond(HttpStatusCode.UnprocessableEntity, ErrorResponse(reason))
            return@post
        }

        val listing = repository.create(req)
        // Publish after commit, not inside the DB transaction -- the write
        // to Postgres is the source of truth; Kafka publish is best-effort
        // async fan-out. If this publish is lost, a periodic reconciliation
        // job (not built here) would replay unpublished rows.
        publisher.publishListingCreated(listing)

        call.respond(HttpStatusCode.Created, listing)
    }

    get("/listings/{id}") {
        val id = call.parameters["id"]!!
        val listing = repository.findById(id)
        if (listing == null) {
            call.respond(HttpStatusCode.NotFound, ErrorResponse("listing $id not found"))
        } else {
            call.respond(listing)
        }
    }

    get("/listings") {
        val limit = call.request.queryParameters["limit"]?.toIntOrNull() ?: 50
        call.respond(repository.listRecent(limit))
    }

    post("/listings/{id}/sold") {
        val id = call.parameters["id"]!!
        val updated = repository.markSold(id)
        if (updated) {
            call.respond(HttpStatusCode.OK)
        } else {
            call.respond(HttpStatusCode.NotFound, ErrorResponse("listing $id not found"))
        }
    }
}
