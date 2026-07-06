package com.skidroad.saga.orchestrator

import com.skidroad.saga.orchestrator.interceptors.*
import com.skidroad.saga.orchestrator.saga.DownstreamClients
import com.skidroad.saga.orchestrator.saga.SagaStateMachine
import com.skidroad.saga.proto.*
import io.grpc.ManagedChannelBuilder
import io.grpc.ServerBuilder
import io.grpc.protobuf.services.ProtoReflectionService
import mu.KotlinLogging

private val log = KotlinLogging.logger {}

fun main() {
    val grpcPort = System.getenv("GRPC_PORT")?.toInt() ?: 50051

    // ── Build downstream gRPC channels ──────────────────────────────────────
    // In production these would be service-mesh addresses (e.g. payment-service:50051)
    val tracingClientInterceptor = TracingClientInterceptor()

    val paymentChannel = ManagedChannelBuilder
        .forAddress(System.getenv("PAYMENT_HOST") ?: "localhost", 50052)
        .usePlaintext()  // TLS in production; use ManagedChannelBuilder.forTarget() with certs
        .build()

    val inventoryChannel = ManagedChannelBuilder
        .forAddress(System.getenv("INVENTORY_HOST") ?: "localhost", 50053)
        .usePlaintext()
        .build()

    val shippingChannel = ManagedChannelBuilder
        .forAddress(System.getenv("SHIPPING_HOST") ?: "localhost", 50054)
        .usePlaintext()
        .build()

    // Attach tracing interceptor to all outbound stubs so trace-id propagates
    val clients = DownstreamClients(
        payment   = PaymentServiceGrpcKt.PaymentServiceCoroutineStub(paymentChannel)
            .withInterceptors(tracingClientInterceptor),
        inventory = InventoryServiceGrpcKt.InventoryServiceCoroutineStub(inventoryChannel)
            .withInterceptors(tracingClientInterceptor),
        shipping  = ShippingServiceGrpcKt.ShippingServiceCoroutineStub(shippingChannel)
            .withInterceptors(tracingClientInterceptor)
    )

    val stateMachine = SagaStateMachine(clients)
    val service      = OrchestratorGrpcService(stateMachine)

    // ── Build gRPC server with interceptor chain ─────────────────────────────
    // Interceptors are applied in reverse registration order by gRPC:
    //   Auth is outermost (first to run), Logging is innermost.
    // Registration order here: Auth, Tracing, Logging
    // Execution order on inbound: Auth → Tracing → Logging → service handler
    val server = ServerBuilder.forPort(grpcPort)
        .addService(service)
        // gRPC Server Reflection — enables grpcurl service discovery
        // grpcurl -plaintext localhost:50051 list
        .addService(ProtoReflectionService.newInstance())
        .intercept(LoggingInterceptor())    // innermost — runs last on way in
        .intercept(TracingInterceptor())
        .intercept(AuthInterceptor(DevTokenValidator()))  // outermost — runs first
        .build()

    server.start()
    log.info { "Saga Orchestrator gRPC server started on port $grpcPort" }
    log.info { "Service reflection enabled — try: grpcurl -plaintext -H 'Authorization: Bearer dev-token' localhost:$grpcPort list" }

    Runtime.getRuntime().addShutdownHook(Thread {
        log.info { "Shutting down gRPC server..." }
        server.shutdown()
        paymentChannel.shutdown()
        inventoryChannel.shutdown()
        shippingChannel.shutdown()
    })

    server.awaitTermination()
}
