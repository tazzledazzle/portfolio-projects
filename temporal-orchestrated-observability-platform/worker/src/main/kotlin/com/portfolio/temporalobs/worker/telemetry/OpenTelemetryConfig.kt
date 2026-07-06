package com.portfolio.temporalobs.worker.telemetry

import io.opentelemetry.api.OpenTelemetry
import io.opentelemetry.api.trace.propagation.W3CTraceContextPropagator
import io.opentelemetry.context.propagation.ContextPropagators
import io.opentelemetry.exporter.otlp.trace.OtlpGrpcSpanExporter
import io.opentelemetry.exporter.prometheus.PrometheusHttpServer
import io.opentelemetry.sdk.OpenTelemetrySdk
import io.opentelemetry.sdk.metrics.SdkMeterProvider
import io.opentelemetry.sdk.metrics.export.MetricReader
import io.opentelemetry.sdk.resources.Resource
import io.opentelemetry.sdk.trace.SdkTracerProvider
import io.opentelemetry.sdk.trace.export.BatchSpanProcessor
import io.opentelemetry.sdk.trace.export.SimpleSpanProcessor
import io.opentelemetry.sdk.trace.export.SpanExporter
import io.opentelemetry.semconv.ServiceAttributes
import java.util.concurrent.TimeUnit

object OpenTelemetryConfig {
    @Volatile
    private var sdk: OpenTelemetrySdk? = null

    @Volatile
    private var prometheusServer: PrometheusHttpServer? = null

    val openTelemetry: OpenTelemetry
        get() = sdk ?: error("OpenTelemetry not initialized — call init() first")

    fun init(
        serviceVersion: String = System.getenv("SERVICE_VERSION") ?: "0.1.0-SNAPSHOT",
        spanExporterOverride: SpanExporter? = null,
        metricReaderOverride: MetricReader? = null,
        enablePrometheus: Boolean = true,
    ): OpenTelemetry {
        if (sdk != null) {
            return sdk!!
        }

        val serviceName = System.getenv("OTEL_SERVICE_NAME") ?: "ai-temporal-worker"
        val resource =
            Resource.getDefault().toBuilder()
                .put(ServiceAttributes.SERVICE_NAME, serviceName)
                .put(ServiceAttributes.SERVICE_VERSION, serviceVersion)
                .build()

        val spanExporter = spanExporterOverride ?: buildOtlpSpanExporter()
        val tracerProviderBuilder =
            SdkTracerProvider.builder()
                .setResource(resource)
        if (spanExporter != null) {
            val processor =
                if (spanExporterOverride != null) {
                    SimpleSpanProcessor.create(spanExporter)
                } else {
                    BatchSpanProcessor.builder(spanExporter).build()
                }
            tracerProviderBuilder.addSpanProcessor(processor)
        }

        val meterProviderBuilder = SdkMeterProvider.builder().setResource(resource)
        if (metricReaderOverride != null) {
            meterProviderBuilder.registerMetricReader(metricReaderOverride)
        } else if (enablePrometheus) {
            val metricsPort = System.getenv("METRICS_PORT")?.toIntOrNull() ?: 9464
            val prometheus =
                PrometheusHttpServer.builder()
                    .setPort(metricsPort)
                    .build()
            prometheusServer = prometheus
            meterProviderBuilder.registerMetricReader(prometheus)
        }

        val otelSdk =
            OpenTelemetrySdk.builder()
                .setTracerProvider(tracerProviderBuilder.build())
                .setMeterProvider(meterProviderBuilder.build())
                .setPropagators(
                    ContextPropagators.create(
                        W3CTraceContextPropagator.getInstance(),
                    ),
                )
                .build()

        sdk = otelSdk
        Metrics.bind(otelSdk)
        return otelSdk
    }

    fun initForTest(
        spanExporter: SpanExporter,
        serviceVersion: String = "test",
    ): OpenTelemetry {
        shutdown()
        System.setProperty("OTEL_TRACES_EXPORTER", "none")
        return init(
            serviceVersion = serviceVersion,
            spanExporterOverride = spanExporter,
            enablePrometheus = false,
        )
    }

    fun shutdown() {
        prometheusServer?.close()
        prometheusServer = null
        sdk?.shutdown()
        sdk = null
        Metrics.unbind()
    }

    private fun buildOtlpSpanExporter(): SpanExporter? {
        if (System.getenv("OTEL_TRACES_EXPORTER")?.equals("none", ignoreCase = true) == true) {
            return null
        }
        val endpoint = System.getenv("OTEL_EXPORTER_OTLP_ENDPOINT") ?: "http://localhost:4317"
        return OtlpGrpcSpanExporter.builder()
            .setEndpoint(endpoint.trimEnd('/'))
            .setTimeout(10, TimeUnit.SECONDS)
            .build()
    }
}
