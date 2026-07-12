package com.marketplace.common.observability

import io.ktor.http.ContentType
import io.ktor.server.application.Application
import io.ktor.server.application.ApplicationCallPipeline
import io.ktor.server.application.call
import io.ktor.server.application.install
import io.ktor.server.metrics.micrometer.MicrometerMetrics
import io.ktor.server.response.respondText
import io.ktor.server.routing.get
import io.ktor.server.routing.routing
import io.micrometer.core.instrument.MeterRegistry
import io.micrometer.core.instrument.Tag
import io.micrometer.core.instrument.distribution.DistributionStatisticConfig
import io.micrometer.prometheusmetrics.PrometheusConfig
import io.micrometer.prometheusmetrics.PrometheusMeterRegistry
import io.ktor.server.request.httpMethod
import io.ktor.server.request.uri
import io.opentelemetry.api.GlobalOpenTelemetry
import io.opentelemetry.api.OpenTelemetry
import io.opentelemetry.api.trace.SpanKind
import io.opentelemetry.api.trace.StatusCode
import io.opentelemetry.context.Context
import io.opentelemetry.context.propagation.TextMapGetter
import io.opentelemetry.sdk.autoconfigure.AutoConfiguredOpenTelemetrySdk
import org.slf4j.MDC
import java.time.Duration

/**
 * Shared Micrometer + OTel wiring for all marketplace services.
 * Metrics: GET /metrics (Prometheus text). Traces: OTLP HTTP → Alloy.
 */
object Observability {
    fun createPrometheusRegistry(serviceName: String): PrometheusMeterRegistry {
        val registry = PrometheusMeterRegistry(PrometheusConfig.DEFAULT)
        registry.config().commonTags(listOf(Tag.of("service", serviceName)))
        return registry
    }

    fun initOpenTelemetry(serviceName: String): OpenTelemetry {
        val endpoint = System.getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
            ?: "http://localhost:4318"
        val name = System.getenv("OTEL_SERVICE_NAME") ?: serviceName

        System.setProperty("otel.service.name", name)
        System.setProperty("otel.exporter.otlp.endpoint", endpoint)
        System.setProperty("otel.exporter.otlp.protocol", "http/protobuf")
        System.setProperty("otel.traces.exporter", "otlp")
        System.setProperty("otel.metrics.exporter", "none")
        System.setProperty("otel.logs.exporter", "none")
        System.setProperty("otel.propagators", "tracecontext,baggage")

        return AutoConfiguredOpenTelemetrySdk.builder()
            .setResultAsGlobal()
            .build()
            .openTelemetrySdk
    }
}

/**
 * Install Micrometer HTTP metrics, Prometheus scrape endpoint, and OTel server spans.
 * Keep CallLogging installed by the service; JSON formatting comes from logback.xml.
 */
fun Application.installObservability(
    serviceName: String,
    registry: PrometheusMeterRegistry,
    openTelemetry: OpenTelemetry = Observability.initOpenTelemetry(serviceName),
) {
    install(MicrometerMetrics) {
        this.registry = registry
        metricName = "http.server.requests"
        distributionStatisticConfig = DistributionStatisticConfig.Builder()
            .percentilesHistogram(true)
            .serviceLevelObjectives(Duration.ofMillis(500).toNanos().toDouble())
            .expiry(Duration.ofMinutes(2))
            .bufferLength(3)
            .build()
    }

    val tracer = openTelemetry.getTracer(serviceName)
    val propagator = openTelemetry.propagators.textMapPropagator
    val headerGetter = object : TextMapGetter<io.ktor.http.Headers> {
        override fun keys(carrier: io.ktor.http.Headers): Iterable<String> = carrier.names()
        override fun get(carrier: io.ktor.http.Headers?, key: String): String? = carrier?.get(key)
    }

    intercept(ApplicationCallPipeline.Monitoring) {
        val parentContext = propagator.extract(Context.current(), call.request.headers, headerGetter)
        val spanName = "${call.request.httpMethod.value} ${call.request.uri}"
        val span = tracer.spanBuilder(spanName)
            .setSpanKind(SpanKind.SERVER)
            .setParent(parentContext)
            .startSpan()

        val scope = span.makeCurrent()
        if (span.spanContext.isValid) {
            MDC.put("trace_id", span.spanContext.traceId)
            MDC.put("span_id", span.spanContext.spanId)
        }
        try {
            proceed()
            val status = call.response.status()?.value ?: 0
            span.setAttribute("http.status_code", status.toLong())
            if (status >= 500) {
                span.setStatus(StatusCode.ERROR)
            }
        } catch (t: Throwable) {
            span.recordException(t)
            span.setStatus(StatusCode.ERROR)
            throw t
        } finally {
            MDC.remove("trace_id")
            MDC.remove("span_id")
            scope.close()
            span.end()
        }
    }

    routing {
        get("/metrics") {
            call.respondText(registry.scrape(), ContentType.Text.Plain)
        }
    }
}

fun MeterRegistry.counterIncrement(name: String, vararg tags: Pair<String, String>) {
    val tagList = tags.map { Tag.of(it.first, it.second) }
    counter(name, tagList).increment()
}

/** No-op helper so services can reference GlobalOpenTelemetry after init. */
fun currentOpenTelemetry(): OpenTelemetry = GlobalOpenTelemetry.get()
