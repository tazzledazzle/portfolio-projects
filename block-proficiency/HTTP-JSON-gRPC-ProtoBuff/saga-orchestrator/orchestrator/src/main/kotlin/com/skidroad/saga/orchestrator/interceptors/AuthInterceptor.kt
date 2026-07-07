package com.skidroad.saga.orchestrator.interceptors

import io.grpc.*
import mu.KotlinLogging

private val log = KotlinLogging.logger {}

/**
 * Server-side interceptor that extracts a Bearer JWT from the `Authorization`
 * metadata key and validates it before allowing the call to proceed.
 *
 * Key concepts demonstrated:
 *  - Reading gRPC metadata (headers) in an interceptor
 *  - Short-circuiting with [ServerCall.close] + [Status.UNAUTHENTICATED]
 *  - Propagating caller identity downstream via [Context]
 */
class AuthInterceptor(
    private val tokenValidator: TokenValidator
) : ServerInterceptor {

    companion object {
        // Context key so downstream handlers can read the validated principal
        val PRINCIPAL_KEY: Context.Key<String> = Context.key("principal")

        private val AUTH_HEADER: Metadata.Key<String> =
            Metadata.Key.of("Authorization", Metadata.ASCII_STRING_MARSHALLER)
    }

    override fun <ReqT : Any, RespT : Any> interceptCall(
        call: ServerCall<ReqT, RespT>,
        headers: Metadata,
        next: ServerCallHandler<ReqT, RespT>
    ): ServerCall.Listener<ReqT> {
        val authHeader = headers.get(AUTH_HEADER)

        if (authHeader == null || !authHeader.startsWith("Bearer ")) {
            log.warn { "Missing or malformed Authorization header" }
            call.close(
                Status.UNAUTHENTICATED.withDescription("Bearer token required"),
                Metadata()
            )
            return object : ServerCall.Listener<ReqT>() {}
        }

        val token = authHeader.removePrefix("Bearer ").trim()
        val principal = tokenValidator.validate(token)

        if (principal == null) {
            log.warn { "Invalid JWT token" }
            call.close(
                Status.UNAUTHENTICATED.withDescription("Invalid or expired token"),
                Metadata()
            )
            return object : ServerCall.Listener<ReqT>() {}
        }

        // Attach validated principal to gRPC Context so any handler can read it
        val ctx = Context.current().withValue(PRINCIPAL_KEY, principal)
        return Contexts.interceptCall(ctx, call, headers, next)
    }
}

/** Pluggable validator — swap in your JWT library (e.g. nimbus-jose-jwt) */
fun interface TokenValidator {
    /** Returns the subject/principal if valid, null if invalid/expired */
    fun validate(token: String): String?
}

/** Simple stub validator for development — replace with real JWT verification */
class DevTokenValidator : TokenValidator {
    override fun validate(token: String): String? =
        if (token == "dev-token") "dev-user" else null
}
