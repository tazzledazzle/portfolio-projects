package com.patterns.grpc

import io.grpc.Server
import io.grpc.ServerBuilder
import io.grpc.ServerInterceptors
import io.grpc.Metadata
import io.grpc.ServerCall
import io.grpc.ServerCallHandler
import io.grpc.ServerInterceptor
import io.grpc.Status
import org.slf4j.LoggerFactory

private val log = LoggerFactory.getLogger("OrderServiceServer")

/**
 * Simple request-logging interceptor. In production this would also:
 * - Extract and propagate trace context (W3C traceparent / B3)
 * - Record request duration as a histogram metric (Prometheus / OTEL)
 * - Validate JWT (if not delegated to the service mesh / gateway)
 */
class LoggingInterceptor : ServerInterceptor {
    override fun <ReqT : Any, RespT : Any> interceptCall(
        call: ServerCall<ReqT, RespT>,
        headers: Metadata,
        next: ServerCallHandler<ReqT, RespT>,
    ): ServerCall.Listener<ReqT> {
        val method = call.methodDescriptor.fullMethodName
        val start = System.currentTimeMillis()
        log.info("gRPC call started: {}", method)
        return object : io.grpc.ForwardingServerCallListener.SimpleForwardingServerCallListener<ReqT>(
            next.startCall(
                object : io.grpc.ForwardingServerCall.SimpleForwardingServerCall<ReqT, RespT>(call) {
                    override fun close(status: Status, trailers: Metadata) {
                        val elapsed = System.currentTimeMillis() - start
                        log.info("gRPC call finished: {} status={} elapsed={}ms", method, status.code, elapsed)
                        super.close(status, trailers)
                    }
                },
                headers,
            )
        ) {}
    }
}

/**
 * Starts the gRPC server on port 50051.
 *
 * Run: gradle run
 * Test with grpcurl: grpcurl -plaintext localhost:50051 list
 */
fun main() {
    val port = System.getenv("GRPC_PORT")?.toIntOrNull() ?: 50051
    val server = buildServer(port)

    Runtime.getRuntime().addShutdownHook(Thread {
        log.info("Shutting down gRPC server…")
        server.shutdown()
        log.info("Server shut down.")
    })

    log.info("Starting OrderService gRPC server on port {}", port)
    server.start()
    log.info("Server started. Awaiting calls.")
    server.awaitTermination()
}

fun buildServer(port: Int): Server {
    val serviceImpl = OrderServiceImpl()
    val interceptedService = ServerInterceptors.intercept(serviceImpl, LoggingInterceptor())

    return ServerBuilder
        .forPort(port)
        .addService(interceptedService)
        .build()
}
