package com.marketplace.listings

import io.ktor.client.call.body
import io.ktor.client.plugins.contentnegotiation.ContentNegotiation
import io.ktor.client.request.get
import io.ktor.client.request.post
import io.ktor.client.request.setBody
import io.ktor.http.ContentType
import io.ktor.http.HttpStatusCode
import io.ktor.http.contentType
import io.ktor.serialization.kotlinx.json.json
import io.ktor.server.testing.testApplication
import io.mockk.every
import io.mockk.mockk
import io.mockk.verify
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Test

class ListingRoutesTest {

    private fun fakeListing(id: String = "listing-1") = Listing(
        id = id,
        sellerId = "seller-1",
        title = "Test Bike",
        description = null,
        priceCents = 5000,
        category = "sporting-goods",
        lat = 47.6062,
        lon = -122.3321,
        status = "ACTIVE",
        createdAtEpochMillis = 1_700_000_000_000L
    )

    @Test
    fun `POST listings returns 201 with listing body and publishes event`() = testApplication {
        val repository = mockk<ListingRepository>()
        val publisher = mockk<EventPublisher>(relaxed = true)
        every { repository.create(any()) } returns fakeListing()

        application { module(repository, publisher) }
        val client = createClient { install(ContentNegotiation) { json() } }

        val response = client.post("/listings") {
            contentType(ContentType.Application.Json)
            setBody(
                CreateListingRequest(
                    sellerId = "seller-1",
                    title = "Test Bike",
                    priceCents = 5000,
                    category = "sporting-goods",
                    lat = 47.6062,
                    lon = -122.3321
                )
            )
        }

        assertEquals(HttpStatusCode.Created, response.status)
        val body = response.body<Listing>()
        assertEquals("listing-1", body.id)
        assertEquals("ACTIVE", body.status)
        verify(exactly = 1) { publisher.publishListingCreated(any()) }
    }

    @Test
    fun `POST listings with blocked keyword returns 422 without touching repository`() = testApplication {
        val repository = mockk<ListingRepository>(relaxed = true)
        val publisher = mockk<EventPublisher>(relaxed = true)

        application { module(repository, publisher) }
        val client = createClient { install(ContentNegotiation) { json() } }

        val response = client.post("/listings") {
            contentType(ContentType.Application.Json)
            setBody(
                CreateListingRequest(
                    sellerId = "seller-1",
                    title = "Stolen bike for sale",
                    priceCents = 5000,
                    category = "sporting-goods",
                    lat = 47.6062,
                    lon = -122.3321
                )
            )
        }

        assertEquals(HttpStatusCode.UnprocessableEntity, response.status)
        verify(exactly = 0) { repository.create(any()) }
        verify(exactly = 0) { publisher.publishListingCreated(any()) }
    }

    @Test
    fun `GET listings id returns 200 with listing body`() = testApplication {
        val repository = mockk<ListingRepository>()
        val publisher = mockk<EventPublisher>(relaxed = true)
        every { repository.findById("listing-1") } returns fakeListing()

        application { module(repository, publisher) }
        val client = createClient { install(ContentNegotiation) { json() } }

        val response = client.get("/listings/listing-1")

        assertEquals(HttpStatusCode.OK, response.status)
        val body = response.body<Listing>()
        assertEquals("listing-1", body.id)
    }

    @Test
    fun `GET listings id returns 404 when listing not found`() = testApplication {
        val repository = mockk<ListingRepository>()
        val publisher = mockk<EventPublisher>(relaxed = true)
        every { repository.findById("unknown") } returns null

        application { module(repository, publisher) }
        val client = createClient { install(ContentNegotiation) { json() } }

        val response = client.get("/listings/unknown")

        assertEquals(HttpStatusCode.NotFound, response.status)
    }
}
