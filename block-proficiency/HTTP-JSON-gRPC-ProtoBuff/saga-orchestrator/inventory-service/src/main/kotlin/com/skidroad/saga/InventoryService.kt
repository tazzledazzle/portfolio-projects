package com.skidroad.saga

import com.skidroad.saga.proto.*
import com.google.protobuf.Duration
import io.grpc.ServerBuilder
import io.grpc.Status
import io.grpc.StatusException
import mu.KotlinLogging
import java.util.UUID
import java.util.concurrent.ConcurrentHashMap

private val log = KotlinLogging.logger {}

/**
 * Inventory Service gRPC implementation.
 *
 * Showcases Protobuf well-known type [Duration] for reservation TTL.
 * A Duration is semantically cleaner than an int64 epoch — it's
 * self-describing and handles negative values correctly.
 */
class InventoryGrpcService : InventoryServiceGrpcKt.InventoryServiceCoroutineImplBase() {

    data class Reservation(val orderId: String, val items: List<OrderItem>, val active: Boolean = true)

    private val reservations   = ConcurrentHashMap<String, Reservation>()
    private val idempotencyMap = ConcurrentHashMap<String, ReserveInventoryResponse>()

    // Simulated stock (sku → quantity)
    private val stock = ConcurrentHashMap(mapOf("SKU-001" to 100, "SKU-002" to 50))

    override suspend fun reserveInventory(request: ReserveInventoryRequest): ReserveInventoryResponse {
        idempotencyMap[request.idempotencyKey]?.let { return it }

        // Validate stock levels
        for (item in request.itemsList) {
            val available = stock.getOrDefault(item.sku, 0)
            if (available < item.quantity) {
                throw StatusException(
                    Status.FAILED_PRECONDITION.withDescription(
                        "Insufficient stock for SKU ${item.sku}: requested=${item.quantity} available=$available"
                    )
                )
            }
        }

        // Deduct stock (in production: use DB transaction)
        request.itemsList.forEach { item ->
            stock.merge(item.sku, -item.quantity, Int::plus)
        }

        val reservationId = UUID.randomUUID().toString()
        reservations[reservationId] = Reservation(request.orderId, request.itemsList)

        val response = reserveInventoryResponse {
            this.reservationId = reservationId
            // Protobuf Duration well-known type: 15 minutes
            reservationTtl = Duration.newBuilder().setSeconds(900).build()
        }

        idempotencyMap[request.idempotencyKey] = response
        log.info { "Inventory reserved: orderId=${request.orderId} reservationId=$reservationId" }
        return response
    }

    override suspend fun releaseInventory(request: ReleaseInventoryRequest): ReleaseInventoryResponse {
        val reservation = reservations[request.reservationId]
            ?: throw StatusException(
                Status.NOT_FOUND.withDescription("Reservation not found: ${request.reservationId}")
            )

        if (!reservation.active) {
            // Idempotent: already released
            return releaseInventoryResponse { released = true }
        }

        // Return stock
        reservation.items.forEach { item ->
            stock.merge(item.sku, item.quantity, Int::plus)
        }
        reservations[request.reservationId] = reservation.copy(active = false)

        log.info { "Inventory released: reservationId=${request.reservationId} reason=${request.reason}" }
        return releaseInventoryResponse { released = true }
    }
}

fun main() {
    val port = System.getenv("GRPC_PORT")?.toInt() ?: 50053
    val server = ServerBuilder.forPort(port)
        .addService(InventoryGrpcService())
        .build()
        .start()
    log.info { "Inventory Service started on port $port" }
    server.awaitTermination()
}
