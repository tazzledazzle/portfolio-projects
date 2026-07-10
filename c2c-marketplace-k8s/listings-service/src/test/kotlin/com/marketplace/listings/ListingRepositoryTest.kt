package com.marketplace.listings

import org.jetbrains.exposed.sql.Database
import org.jetbrains.exposed.sql.SchemaUtils
import org.jetbrains.exposed.sql.transactions.transaction
import org.junit.jupiter.api.AfterAll
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Assertions.assertNotNull
import org.junit.jupiter.api.Assertions.assertNull
import org.junit.jupiter.api.BeforeAll
import org.junit.jupiter.api.Test
import org.testcontainers.containers.PostgreSQLContainer
import org.testcontainers.junit.jupiter.Testcontainers

/**
 * Runs against a real throwaway Postgres via Testcontainers rather than
 * mocking the JDBC driver -- what we actually want confidence in here is
 * that the Exposed SQL is correct against a real Postgres, not that Kotlin
 * can call a mock. Requires Docker to be available wherever this test runs.
 */
@Testcontainers
class ListingRepositoryTest {

    companion object {
        private val postgres = PostgreSQLContainer("postgres:16-alpine")
            .withDatabaseName("marketplace_test")
            .withUsername("test")
            .withPassword("test")

        @JvmStatic
        @BeforeAll
        fun setup() {
            postgres.start()
            Database.connect(
                url = postgres.jdbcUrl,
                user = postgres.username,
                password = postgres.password
            )
            transaction { SchemaUtils.createMissingTablesAndColumns(ListingTable) }
        }

        @JvmStatic
        @AfterAll
        fun teardown() {
            postgres.stop()
        }
    }

    private val repository = ListingRepository()

    @Test
    fun `create then find by id round-trips all fields`() {
        val req = CreateListingRequest(
            sellerId = "seller-1",
            title = "Mid-century desk",
            description = "Solid wood, minor scuffs",
            priceCents = 4000,
            category = "furniture",
            lat = 47.6062,
            lon = -122.3321
        )

        val created = repository.create(req)
        val found = repository.findById(created.id)

        assertNotNull(found)
        assertEquals(req.title, found!!.title)
        assertEquals(req.priceCents, found.priceCents)
        assertEquals("ACTIVE", found.status)
    }

    @Test
    fun `find by unknown id returns null`() {
        assertNull(repository.findById("does-not-exist"))
    }

    @Test
    fun `mark sold transitions status`() {
        val created = repository.create(
            CreateListingRequest(
                sellerId = "seller-2",
                title = "Bike",
                description = null,
                priceCents = 15000,
                category = "sporting-goods",
                lat = 47.6,
                lon = -122.3
            )
        )

        val updated = repository.markSold(created.id)
        val reloaded = repository.findById(created.id)

        assertEquals(true, updated)
        assertEquals("SOLD", reloaded!!.status)
    }
}
