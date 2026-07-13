package com.marketplace.synth

import io.ktor.client.HttpClient
import io.ktor.client.call.body
import io.ktor.client.engine.cio.CIO
import io.ktor.client.plugins.HttpTimeout
import io.ktor.client.plugins.contentnegotiation.ContentNegotiation
import io.ktor.client.request.get
import io.ktor.client.request.parameter
import io.ktor.client.request.post
import io.ktor.client.request.setBody
import io.ktor.client.statement.bodyAsText
import io.ktor.http.ContentType
import io.ktor.http.contentType
import io.ktor.http.isSuccess
import io.ktor.serialization.kotlinx.json.json
import java.io.Closeable
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json

class MarketplaceHttpException(
    val status: Int,
    val responseBody: String,
    message: String = "HTTP $status: $responseBody"
) : Exception(message)

@Serializable
data class CreateListingBody(
    val sellerId: String,
    val title: String,
    val description: String? = null,
    val priceCents: Int,
    val category: String,
    val lat: Double,
    val lon: Double
)

@Serializable
data class ListingResponse(
    val id: String,
    val title: String = "",
    val sellerId: String = "",
    val priceCents: Int = 0,
    val category: String = "",
    val status: String = ""
)

@Serializable
data class CreateOrderBody(
    val listingId: String,
    val buyerId: String,
    val sellerId: String,
    val amountCents: Int
)

@Serializable
data class OrderResponse(
    val id: String,
    val status: String = ""
)

@Serializable
data class EscrowActionResponse(
    val orderId: String,
    val status: String
)

@Serializable
data class SearchHit(
    val listingId: String,
    val title: String,
    val priceCents: Int = 0,
    val category: String = "",
    val distanceKm: Double? = null
)

class MarketplaceClient(
    listingsUrl: String,
    searchUrl: String,
    paymentsUrl: String,
    private val http: HttpClient = defaultHttpClient()
) : Closeable {
    private val listingsBase = listingsUrl.trimEnd('/')
    private val searchBase = searchUrl.trimEnd('/')
    private val paymentsBase = paymentsUrl.trimEnd('/')

    suspend fun createListing(
        sellerId: String,
        title: String,
        description: String?,
        priceCents: Int,
        category: String,
        lat: Double,
        lon: Double
    ): String {
        val response = http.post("$listingsBase/listings") {
            contentType(ContentType.Application.Json)
            setBody(
                CreateListingBody(
                    sellerId = sellerId,
                    title = title,
                    description = description,
                    priceCents = priceCents,
                    category = category,
                    lat = lat,
                    lon = lon
                )
            )
        }
        if (!response.status.isSuccess()) {
            throw MarketplaceHttpException(response.status.value, response.bodyAsText())
        }
        return response.body<ListingResponse>().id
    }

    suspend fun search(q: String, lat: Double, lon: Double, radiusKm: Int = 25): List<SearchHit> {
        val response = http.get("$searchBase/search") {
            parameter("q", q)
            parameter("lat", lat)
            parameter("lon", lon)
            parameter("radiusKm", radiusKm)
        }
        if (!response.status.isSuccess()) {
            throw MarketplaceHttpException(response.status.value, response.bodyAsText())
        }
        return response.body()
    }

    suspend fun createOrder(
        listingId: String,
        buyerId: String,
        sellerId: String,
        amountCents: Int
    ): String {
        val response = http.post("$paymentsBase/orders") {
            contentType(ContentType.Application.Json)
            setBody(
                CreateOrderBody(
                    listingId = listingId,
                    buyerId = buyerId,
                    sellerId = sellerId,
                    amountCents = amountCents
                )
            )
        }
        if (!response.status.isSuccess()) {
            throw MarketplaceHttpException(response.status.value, response.bodyAsText())
        }
        return response.body<OrderResponse>().id
    }

    suspend fun confirmDelivery(orderId: String): String {
        val response = http.post("$paymentsBase/orders/$orderId/confirm-delivery")
        if (!response.status.isSuccess()) {
            throw MarketplaceHttpException(response.status.value, response.bodyAsText())
        }
        return response.body<EscrowActionResponse>().status
    }

    suspend fun dispute(orderId: String): String {
        val response = http.post("$paymentsBase/orders/$orderId/dispute")
        if (!response.status.isSuccess()) {
            throw MarketplaceHttpException(response.status.value, response.bodyAsText())
        }
        return response.body<EscrowActionResponse>().status
    }

    override fun close() {
        http.close()
    }

    companion object {
        fun defaultHttpClient(): HttpClient = HttpClient(CIO) {
            install(ContentNegotiation) {
                json(Json { ignoreUnknownKeys = true })
            }
            install(HttpTimeout) {
                connectTimeoutMillis = 5_000
                requestTimeoutMillis = 15_000
            }
        }
    }
}
