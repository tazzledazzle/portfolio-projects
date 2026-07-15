package com.marketplace.payments

import org.jetbrains.exposed.sql.Database
import org.jetbrains.exposed.sql.SchemaUtils
import org.jetbrains.exposed.sql.transactions.transaction
import org.junit.jupiter.api.AfterAll
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Assertions.assertNotNull
import org.junit.jupiter.api.Assertions.assertThrows
import org.junit.jupiter.api.BeforeAll
import org.junit.jupiter.api.Test
import org.testcontainers.containers.PostgreSQLContainer
import org.testcontainers.junit.jupiter.Testcontainers

@Testcontainers
class OrderRepositoryTest {

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
            transaction { SchemaUtils.createMissingTablesAndColumns(OrderTable, EscrowHoldTable) }
        }

        @JvmStatic
        @AfterAll
        fun teardown() {
            postgres.stop()
        }
    }

    private val repository = OrderRepository()

    private fun sampleRequest() = CreateOrderRequest(
        listingId = "listing-1",
        buyerId = "buyer-1",
        sellerId = "seller-1",
        amountCents = 5000
    )

    @Test
    fun `createWithHold persists order and escrow hold atomically`() {
        val order = repository.createWithHold(sampleRequest())

        assertNotNull(order.id)
        assertEquals("HELD", order.status)
        assertEquals(EscrowStatus.HELD, repository.currentEscrowStatus(order.id))
    }

    @Test
    fun `applyEvent ConfirmDelivery transitions HELD to RELEASED`() {
        val order = repository.createWithHold(sampleRequest())

        val result = repository.applyEvent(order.id, EscrowEvent.ConfirmDelivery)

        assertEquals(EscrowStatus.RELEASED, result)
        assertEquals(EscrowStatus.RELEASED, repository.currentEscrowStatus(order.id))
    }

    @Test
    fun `applyEvent BuyerDispute transitions HELD to REFUNDED`() {
        val order = repository.createWithHold(sampleRequest())

        val result = repository.applyEvent(order.id, EscrowEvent.BuyerDispute)

        assertEquals(EscrowStatus.REFUNDED, result)
        assertEquals(EscrowStatus.REFUNDED, repository.currentEscrowStatus(order.id))
    }

    @Test
    fun `applyEvent on non-existent order throws NoSuchElementException`() {
        assertThrows(NoSuchElementException::class.java) {
            repository.applyEvent("does-not-exist", EscrowEvent.ConfirmDelivery)
        }
    }

    @Test
    fun `applyEvent on already RELEASED order throws IllegalEscrowTransitionException`() {
        val order = repository.createWithHold(sampleRequest())
        repository.applyEvent(order.id, EscrowEvent.ConfirmDelivery) // HELD -> RELEASED

        assertThrows(IllegalEscrowTransitionException::class.java) {
            repository.applyEvent(order.id, EscrowEvent.ConfirmDelivery) // RELEASED -> illegal
        }
    }
}
