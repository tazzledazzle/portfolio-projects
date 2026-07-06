package com.skidroad.saga.orchestrator.interceptors

import io.grpc.*
import mu.KotlinLogging

private val log = KotlinLogging.logger {}

/**
 * Server-side interceptor that logs every inbound gRPC call with:
 *  - Method name
 *  - Trace ID (from context, set by [TracingInterceptor] which must run first)
 *  - Principal (from context, set by [AuthInterceptor] which must run first)
 *  - Call duration (ms)
 *  - Final gRPC status code
 *
 * Interceptor ordering matters: register Auth → Tracing → Logging on the
 * ServerBuilder. gRPC applies them outermost-first on the way in.
 */
class LoggingInterceptor : ServerInterceptor {

    override fun <ReqT : Any, RespT : Any> interceptCall(
        call: ServerCall<ReqT, RespT>,
        headers: Metadata,
        next: ServerCallHandler<ReqT, RespT>
    ): ServerCall.Listener<ReqT> {
        val method    = call.methodDescriptor.fullMethodName
        val traceId   = TracingInterceptor.TRACE_ID_KEY.get() ?: "no-trace"
        val principal = AuthInterceptor.PRINCIPAL_KEY.get()   ?: "anonymous"
        val startMs   = System.currentTimeMillis()

        log.info { "→ gRPC $method traceId=$traceId principal=$principal" }

        // Wrap the ServerCall to intercept the close event (captures final status)
        val loggingCall = object : ForwardingServerCall.SimpleForwardingServerCall<ReqT, RespT>(call) {
            override fun close(status: Status, trailers: Metadata) {
                val durationMs = System.currentTimeMillis() - startMs
                if (status.isOk) {
                    log.info { "← gRPC $method OK ${durationMs}ms traceId=$traceId" }
                } else {
                    log.warn { "← gRPC $method ${status.code} '${status.description}' ${durationMs}ms traceId=$traceId" }
                }
                super.close(status, trailers)
            }
        }

        return next.startCall(loggingCall, headers)
    }
}
