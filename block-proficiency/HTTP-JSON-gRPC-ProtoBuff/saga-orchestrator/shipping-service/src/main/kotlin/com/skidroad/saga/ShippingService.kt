package com.skidroad.saga

import com.skidroad.saga.proto.*
import com.google.protobuf.Timestamp
import io.grpc.ServerBuilder
import io.grpc.Status
import io.grpc.StatusException
import mu.KotlinLogging
import java.time.Instant
import java.time.temporal.ChronoUnit
import java.util.UUID
import java.util.concurrent.ConcurrentHashMap

private val log = KotlinLogging.logger {}

/**
 * Shipping Service gRPC implementation.
 *
 * Set SHIPPING_FAIL_RATE env var (0.0–1.0) to control failure probability.
 * Default: 0.3 (30% failure rate) — enough to trigger compensation in demos.
 *
 * This is where the saga compensation gets exercised: when this service
 * returns UNAVAILABLE, the orchestrator's CompensationEngine will roll back
 * the inventory reservation and payment charge.
 */
class ShippingGrpcService : ShippingServiceGrpcKt.ShippingServiceCoroutineImplBase() {

    private val shipments      = ConcurrentHashMap<String, CreateShipmentResponse>()
    private val idempotencyMap = ConcurrentHashMap<String, CreateShipmentResponse>()

    private val failRate: Double = System.getenv("SHIPPING_FAIL_RATE")?.toDouble() ?: 0.3

    override suspend fun createShipment(request: CreateShipmentRequest): CreateShipmentResponse {
        idempotencyMap[request.idempotencyKey]?.let { return it }

        // Simulate transient unavailability
        if (Math.random() < failRate) {
            throw StatusException(
                Status.UNAVAILABLE.withDescription("Shipping carrier API unavailable (simulated)")
            )
        }

        val shipmentId     = "SHIP-${UUID.randomUUID().toString().take(8).uppercase()}"
        val trackingNumber = "1Z${UUID.randomUUID().toString().replace("-","").take(16).uppercase()}"

        val response = createShipmentResponse {
            this.shipmentId     = shipmentId
            this.trackingNumber = trackingNumber
            // Protobuf Timestamp well-known type: estimated delivery in 3 days
            estimatedDelivery   = Instant.now().plus(3, ChronoUnit.DAYS).toProtoTimestamp()
        }

        shipments[shipmentId] = response
        idempotencyMap[request.idempotencyKey] = response

        log.info { "Shipment created: orderId=${request.orderId} shipmentId=$shipmentId tracking=$trackingNumber" }
        return response
    }
}

fun main() {
    val port = System.getenv("GRPC_PORT")?.toInt() ?: 50054
    val server = ServerBuilder.forPort(port)
        .addService(ShippingGrpcService())
        .build()
        .start()
    log.info { "Shipping Service started on port $port (fail_rate=${System.getenv("SHIPPING_FAIL_RATE") ?: "0.3"})" }
    server.awaitTermination()
}

private fun Instant.toProtoTimestamp(): Timestamp =
    Timestamp.newBuilder().setSeconds(epochSecond).setNanos(nano).build()
