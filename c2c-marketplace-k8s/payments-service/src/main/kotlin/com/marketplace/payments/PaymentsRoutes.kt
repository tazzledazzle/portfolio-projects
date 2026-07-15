package com.marketplace.payments

import io.ktor.http.HttpStatusCode
import io.ktor.server.application.call
import io.ktor.server.request.receive
import io.ktor.server.response.respond
import io.ktor.server.routing.Route
import io.ktor.server.routing.post
import io.micrometer.core.instrument.MeterRegistry
import io.micrometer.core.instrument.Tag
import kotlinx.serialization.Serializable

@Serializable
data class ErrorResponse(val error: String)

fun Route.paymentsRoutes(
    repository: OrderRepository,
    publisher: EventPublisher,
    meters: MeterRegistry,
) {

    post("/orders") {
        val req = call.receive<CreateOrderRequest>()
        val order = repository.createWithHold(req)
        publisher.publishOrderCreated(order)
        call.respond(HttpStatusCode.Created, order)
    }

    post("/orders/{id}/confirm-delivery") {
        val id = call.parameters["id"]!!
        handleEscrowEvent(call, repository, publisher, meters, id, EscrowEvent.ConfirmDelivery)
    }

    post("/orders/{id}/dispute") {
        val id = call.parameters["id"]!!
        handleEscrowEvent(call, repository, publisher, meters, id, EscrowEvent.BuyerDispute)
    }
}

private suspend fun handleEscrowEvent(
    call: io.ktor.server.application.ApplicationCall,
    repository: OrderRepository,
    publisher: EventPublisher,
    meters: MeterRegistry,
    orderId: String,
    event: EscrowEvent
) {
    val escrowCounter = { result: String ->
        meters.counter("escrow_transitions", listOf(Tag.of("result", result)))
    }
    try {
        val newStatus = repository.applyEvent(orderId, event)
        if (newStatus == EscrowStatus.RELEASED) {
            publisher.publishOrderCompleted(orderId)
        }
        escrowCounter("success").increment()
        call.respond(HttpStatusCode.OK, mapOf("orderId" to orderId, "status" to newStatus.name))
    } catch (e: IllegalEscrowTransitionException) {
        // Expected client error (409) — does not burn availability / escrow SLI budget
        escrowCounter("client_error").increment()
        call.respond(HttpStatusCode.Conflict, ErrorResponse(e.message ?: "illegal transition"))
    } catch (e: NoSuchElementException) {
        escrowCounter("client_error").increment()
        call.respond(HttpStatusCode.NotFound, ErrorResponse(e.message ?: "order not found"))
    } catch (e: Exception) {
        escrowCounter("server_error").increment()
        throw e
    }
}
