package com.marketplace.search

import com.marketplace.common.ListingCreatedEvent
import io.ktor.client.HttpClient
import io.ktor.client.engine.cio.CIO
import io.ktor.client.request.put
import io.ktor.client.request.post
import io.ktor.client.request.setBody
import io.ktor.client.statement.bodyAsText
import io.ktor.http.ContentType
import io.ktor.http.contentType
import kotlinx.serialization.Serializable
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.jsonArray
import kotlinx.serialization.json.jsonObject
import kotlinx.serialization.json.jsonPrimitive
import org.slf4j.LoggerFactory

const val LISTINGS_INDEX = "listings"

@Serializable
data class ListingDocument(
    val listingId: String,
    val sellerId: String,
    val title: String,
    val description: String?,
    val priceCents: Int,
    val category: String,
    // OpenSearch geo_point expects "lat,lon" as a string or {lat, lon} object;
    // we use the object form for clarity.
    val location: GeoPoint,
    val status: String,
    val createdAt: Long
)

@Serializable
data class GeoPoint(val lat: Double, val lon: Double)

@Serializable
data class SearchResult(
    val listingId: String,
    val title: String,
    val priceCents: Int,
    val category: String,
    val distanceKm: Double? = null
)

/**
 * Thin wrapper over OpenSearch's REST API using plain HTTP calls instead of
 * the full OpenSearch Java client -- for a service that only ever does
 * "index a document" and "run one query shape", the official client is a
 * lot of dependency weight for not much benefit. If query complexity grows
 * (aggregations, multiple query types), switching to the real client pays
 * for itself.
 */
class OpenSearchClient(private val baseUrl: String) {
    private val logger = LoggerFactory.getLogger(javaClass)
    private val json = Json { ignoreUnknownKeys = true }
    private val http = HttpClient(CIO)

    suspend fun ensureIndex() {
        val mapping = """
            {
              "mappings": {
                "properties": {
                  "title": { "type": "text" },
                  "description": { "type": "text" },
                  "category": { "type": "keyword" },
                  "priceCents": { "type": "integer" },
                  "status": { "type": "keyword" },
                  "location": { "type": "geo_point" }
                }
              }
            }
        """.trimIndent()
        try {
            http.put("$baseUrl/$LISTINGS_INDEX") {
                contentType(ContentType.Application.Json)
                setBody(mapping)
            }
        } catch (e: Exception) {
            // Index probably already exists -- fine to ignore on restart.
            logger.info("ensureIndex: ${e.message}")
        }
    }

    suspend fun indexListing(event: ListingCreatedEvent) {
        val doc = ListingDocument(
            listingId = event.listingId,
            sellerId = event.sellerId,
            title = event.title,
            description = event.description,
            priceCents = event.priceCents,
            category = event.category,
            location = GeoPoint(event.lat, event.lon),
            status = "ACTIVE",
            createdAt = event.createdAt
        )
        val body = json.encodeToString(doc)
        http.put("$baseUrl/$LISTINGS_INDEX/_doc/${event.listingId}") {
            contentType(ContentType.Application.Json)
            setBody(body)
        }
        logger.info("Indexed listing ${event.listingId}")
    }

    /**
     * BM25 text relevance combined with a geo-distance filter -- the two
     * signals a browse/search screen actually needs: "is this a good match
     * for the query" and "is this close enough to matter to the buyer".
     */
    suspend fun search(query: String, lat: Double?, lon: Double?, radiusKm: Int): List<SearchResult> {
        val geoFilter = if (lat != null && lon != null) {
            """
            , "filter": [
                {
                  "geo_distance": {
                    "distance": "${radiusKm}km",
                    "location": { "lat": $lat, "lon": $lon }
                  }
                }
              ]
            """.trimIndent()
        } else ""

        val body = """
            {
              "query": {
                "bool": {
                  "must": [
                    { "multi_match": { "query": ${Json.encodeToString(query)}, "fields": ["title^2", "description"] } }
                  ],
                  "filter": [ { "term": { "status": "ACTIVE" } } ]
                  $geoFilter
                }
              },
              "size": 30
            }
        """.trimIndent()

        val response = http.post("$baseUrl/$LISTINGS_INDEX/_search") {
            contentType(ContentType.Application.Json)
            setBody(body)
        }.bodyAsText()

        return parseHits(response)
    }

    private fun parseHits(responseJson: String): List<SearchResult> {
        return try {
            val root = json.parseToJsonElement(responseJson).jsonObject
            val hits = root["hits"]?.jsonObject?.get("hits")?.jsonArray ?: return emptyList()
            hits.mapNotNull { hit ->
                val source = hit.jsonObject["_source"]?.jsonObject ?: return@mapNotNull null
                SearchResult(
                    listingId = source["listingId"]?.jsonPrimitive?.content ?: return@mapNotNull null,
                    title = source["title"]?.jsonPrimitive?.content ?: "",
                    priceCents = source["priceCents"]?.jsonPrimitive?.content?.toIntOrNull() ?: 0,
                    category = source["category"]?.jsonPrimitive?.content ?: ""
                )
            }
        } catch (e: Exception) {
            logger.warn("Failed to parse OpenSearch response: ${e.message}")
            emptyList()
        }
    }
}
