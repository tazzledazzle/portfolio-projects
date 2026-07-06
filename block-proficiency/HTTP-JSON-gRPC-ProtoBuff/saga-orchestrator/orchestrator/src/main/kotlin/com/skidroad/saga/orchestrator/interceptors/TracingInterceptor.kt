package com.skidroad.saga.orchestrator.interceptors

import io.grpc.*
import mu.KotlinLogging
import java.util.UUID

private val log = KotlinLogging.logger {}

/**
 * Server-side interceptor that implements W3C Trace Context (traceparent header).
 *
 * Key concepts demonstrated:
 *  - Reading + writing gRPC metadata for distributed tracing
 *  - Generating a trace-id when none is present (acting as a trace root)
 *  - Storing trace context in gRPC [Context] for downstream propagation
 *  - When making outbound gRPC calls, [TracingClientInterceptor] reads this
 *    context and injects it into outbound metadata so the trace continues.
 *
 * W3C traceparent format: 00-<trace-id>-<parent-id>-<flags>
 * See: https://www.w3.org/TR/trace-context/
 */
class TracingInterceptor : ServerInterceptor {

    companion object {
        val TRACE_ID_KEY: Context.Key<String> = Context.key("trace-id")
        val SPAN_ID_KEY:  Context.Key<String> = Context.key("span-id")

        private val TRACEPARENT: Metadata.Key<String> =
            Metadata.Key.of("traceparent", Metadata.ASCII_STRING_MARSHALLER)
    }

    override fun <ReqT : Any, RespT : Any> interceptCall(
        call: ServerCall<ReqT, RespT>,
        headers: Metadata,
        next: ServerCallHandler<ReqT, RespT>
    ): ServerCall.Listener<ReqT> {
        val (traceId, spanId) = parseOrGenerate(headers.get(TRACEPARENT))

        log.debug { "Trace context: traceId=$traceId spanId=$spanId method=${call.methodDescriptor.fullMethodName}" }

        val ctx = Context.current()
            .withValue(TRACE_ID_KEY, traceId)
            .withValue(SPAN_ID_KEY, spanId)

        return Contexts.interceptCall(ctx, call, headers, next)
    }

    private fun parseOrGenerate(traceparent: String?): Pair<String, String> {
        // Expected format: 00-{32hex}-{16hex}-{2hex}
        return if (traceparent != null) {
            val parts = traceparent.split("-")
            if (parts.size == 4) {
                val traceId = parts[1]
                val newSpanId = UUID.randomUUID().toString().replace("-", "").take(16)
                Pair(traceId, newSpanId)
            } else {
                generateNew()
            }
        } else {
            generateNew()
        }
    }

    private fun generateNew(): Pair<String, String> {
        val traceId = UUID.randomUUID().toString().replace("-", "")
        val spanId  = UUID.randomUUID().toString().replace("-", "").take(16)
        return Pair(traceId, spanId)
    }
}

/**
 * Client-side interceptor that injects the current trace context into
 * outbound gRPC calls. Attach this to every downstream stub so that
 * payment/inventory/shipping services receive the trace header.
 *
 * Usage:
 *   val stub = PaymentServiceGrpcKt.PaymentServiceCoroutineStub(channel)
 *       .withInterceptors(TracingClientInterceptor())
 */
class TracingClientInterceptor : ClientInterceptor {

    companion object {
        private val TRACEPARENT: Metadata.Key<String> =
            Metadata.Key.of("traceparent", Metadata.ASCII_STRING_MARSHALLER)
    }

    override fun <ReqT : Any, RespT : Any> interceptCall(
        method: MethodDescriptor<ReqT, RespT>,
        callOptions: CallOptions,
        next: Channel
    ): ClientCall<ReqT, RespT> {
        val traceId = TracingInterceptor.TRACE_ID_KEY.get() ?: return next.newCall(method, callOptions)
        val spanId  = TracingInterceptor.SPAN_ID_KEY.get()  ?: return next.newCall(method, callOptions)

        return object : ForwardingClientCall.SimpleForwardingClientCall<ReqT, RespT>(
            next.newCall(method, callOptions)
        ) {
            override fun start(responseListener: Listener<RespT>, headers: Metadata) {
                // Inject W3C traceparent into outbound metadata
                val traceparent = "00-$traceId-$spanId-01"
                headers.put(TRACEPARENT, traceparent)
                super.start(responseListener, headers)
            }
        }
    }
}
