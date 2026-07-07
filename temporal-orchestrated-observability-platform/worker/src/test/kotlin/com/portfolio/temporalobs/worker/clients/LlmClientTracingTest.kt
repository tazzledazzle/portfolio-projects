package com.portfolio.temporalobs.worker.clients

import com.portfolio.temporalobs.worker.telemetry.Metrics
import com.portfolio.temporalobs.worker.telemetry.OpenTelemetryConfig
import io.opentelemetry.api.trace.SpanKind
import io.opentelemetry.api.trace.Tracer
import io.opentelemetry.sdk.testing.exporter.InMemorySpanExporter
import org.junit.jupiter.api.Assertions.assertTrue
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.Assertions.assertFalse
import org.junit.jupiter.api.Assertions.assertNotNull
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test

class LlmClientTracingTest {
    private lateinit var spanExporter: InMemorySpanExporter
    private lateinit var server: MockWebServer

    @BeforeEach
    fun setUp() {
        spanExporter = InMemorySpanExporter.create()
        OpenTelemetryConfig.initForTest(spanExporter)
        server = MockWebServer()
        server.start()
        server.enqueue(
            MockResponse()
                .setBody(
                    """
                    {
                      "text": "answer",
                      "citation": "doc-1"
                    }
                    """.trimIndent(),
                )
                .addHeader("Content-Type", "application/json"),
        )
    }

    @AfterEach
    fun tearDown() {
        server.shutdown()
        OpenTelemetryConfig.shutdown()
    }

    @Test
    fun llmRequestCreatesChildSpanAndTraceparent() {
        val otel = OpenTelemetryConfig.openTelemetry
        val tracer: Tracer = otel.getTracer("test")
        val parentSpan = tracer.spanBuilder("activity.llmComplete").setSpanKind(SpanKind.INTERNAL).startSpan()

        parentSpan.makeCurrent().use {
            val client =
                LlmClient(
                    baseUrl = server.url("/").toString().trimEnd('/'),
                    callFactory = LlmClient.instrumentedCallFactory(otel),
                )
            val result = client.complete("What is OTel?", listOf("chunk-a"))
            assertFalse(result.text.isBlank())
        }
        parentSpan.end()

        val recordedRequest = server.takeRequest()
        val traceparent = recordedRequest.getHeader("traceparent")
        assertNotNull(traceparent)
        assertFalse(traceparent!!.isBlank())

        val childSpans =
            spanExporter.finishedSpanItems.filter { span ->
                span.name != "activity.llmComplete"
            }
        assertTrue(
            childSpans.isNotEmpty(),
            "expected OkHttp client span, got: ${spanExporter.finishedSpanItems.map { it.name }}",
        )
        Metrics.recordLlmDuration("ok", 0.01)
    }
}
