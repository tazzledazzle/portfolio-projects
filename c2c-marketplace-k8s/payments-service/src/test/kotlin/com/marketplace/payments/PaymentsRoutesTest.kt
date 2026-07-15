package com.marketplace.payments

import io.ktor.client.call.body
import io.ktor.client.plugins.contentnegotiation.ContentNegotiation
import io.ktor.client.request.post
import io.ktor.client.request.setBody
import io.ktor.http.ContentType
import io.ktor.http.HttpStatusCode
import io.ktor.http.contentType
import io.ktor.serialization.kotlinx.json.json
import io.ktor.server.testing.testApplication
import io.mockk.every
import io.mockk.mockk
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Test

class PaymentsRoutesTest {

    private fun makeOrder(id: String = "order-1", status: String = "HELD") =
        Order(
            id = id,
            listingId = "listing-1",
            buyerId = "buyer-1",
            sellerId = "seller-1",
            amountCents = 5000,
            status = status
        )

    @Test
    fun `POST orders returns 201 with orderId and status HELD`() = testApplication {
        val repository = mockk<OrderRepository>()
        val publisher = mockk<EventPublisher>(relaxed = true)
        every { repository.createWithHold(any()) } returns makeOrder()

        application { module(repository, publisher) }
        val client = createClient { install(ContentNegotiation) { json() } }

        val response = client.post("/orders") {
            contentType(ContentType.Application.Json)
            setBody(CreateOrderRequest("listing-1", "buyer-1", "seller-1", 5000))
        }

        assertEquals(HttpStatusCode.Created, response.status)
        val body = response.body<Order>()
        assertEquals("order-1", body.id)
        assertEquals("HELD", body.status)
    }

    @Test
    fun `POST orders id confirm-delivery returns 200 with status RELEASED`() = testApplication {
        val repository = mockk<OrderRepository>()
        val publisher = mockk<EventPublisher>(relaxed = true)
        every { repository.applyEvent("order-1", EscrowEvent.ConfirmDelivery) } returns EscrowStatus.RELEASED

        application { module(repository, publisher) }
        val client = createClient { install(ContentNegotiation) { json() } }

        val response = client.post("/orders/order-1/confirm-delivery")

        assertEquals(HttpStatusCode.OK, response.status)
        val body = response.body<Map<String, String>>()
        assertEquals("RELEASED", body["status"])
    }

    @Test
    fun `POST orders id dispute returns 200 with status REFUNDED`() = testApplication {
        val repository = mockk<OrderRepository>()
        val publisher = mockk<EventPublisher>(relaxed = true)
        every { repository.applyEvent("order-1", EscrowEvent.BuyerDispute) } returns EscrowStatus.REFUNDED

        application { module(repository, publisher) }
        val client = createClient { install(ContentNegotiation) { json() } }

        val response = client.post("/orders/order-1/dispute")

        assertEquals(HttpStatusCode.OK, response.status)
        val body = response.body<Map<String, String>>()
        assertEquals("REFUNDED", body["status"])
    }

    @Test
    fun `POST orders id confirm-delivery returns 409 when order already released`() = testApplication {
        val repository = mockk<OrderRepository>()
        val publisher = mockk<EventPublisher>(relaxed = true)
        every { repository.applyEvent(any(), EscrowEvent.ConfirmDelivery) } throws
            IllegalEscrowTransitionException("cannot apply ConfirmDelivery to order already in state RELEASED")

        application { module(repository, publisher) }
        val client = createClient { install(ContentNegotiation) { json() } }

        val response = client.post("/orders/order-1/confirm-delivery")

        assertEquals(HttpStatusCode.Conflict, response.status)
    }

    @Test
    fun `POST orders id confirm-delivery returns 404 when order not found`() = testApplication {
        val repository = mockk<OrderRepository>()
        val publisher = mockk<EventPublisher>(relaxed = true)
        every { repository.applyEvent(any(), EscrowEvent.ConfirmDelivery) } throws
            NoSuchElementException("no escrow hold for order unknown")

        application { module(repository, publisher) }
        val client = createClient { install(ContentNegotiation) { json() } }

        val response = client.post("/orders/unknown/confirm-delivery")

        assertEquals(HttpStatusCode.NotFound, response.status)
    }
}
