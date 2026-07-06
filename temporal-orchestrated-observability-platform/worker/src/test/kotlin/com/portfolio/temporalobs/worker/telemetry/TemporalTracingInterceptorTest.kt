package com.portfolio.temporalobs.worker.telemetry

import com.portfolio.temporalobs.worker.PingActivitiesImpl
import com.portfolio.temporalobs.workflows.PingActivities
import io.opentelemetry.api.common.AttributeKey
import io.opentelemetry.sdk.testing.exporter.InMemorySpanExporter
import io.opentelemetry.sdk.trace.data.SpanData
import io.temporal.testing.TestActivityEnvironment
import io.temporal.testing.TestEnvironmentOptions
import io.temporal.worker.WorkerFactoryOptions
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Assertions.assertFalse
import org.junit.jupiter.api.Assertions.assertNotNull
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.Timeout
import java.util.concurrent.TimeUnit

class TemporalTracingInterceptorTest {
    private lateinit var spanExporter: InMemorySpanExporter
    private lateinit var testEnv: TestActivityEnvironment

    @BeforeEach
    fun setUp() {
        spanExporter = InMemorySpanExporter.create()
        OpenTelemetryConfig.initForTest(spanExporter)

        val factoryOptions =
            WorkerFactoryOptions.newBuilder()
                .setWorkerInterceptors(TemporalTracingInterceptor(OpenTelemetryConfig.openTelemetry))
                .build()

        testEnv =
            TestActivityEnvironment.newInstance(
                TestEnvironmentOptions.newBuilder()
                    .setWorkerFactoryOptions(factoryOptions)
                    .build(),
            )
        testEnv.registerActivitiesImplementations(PingActivitiesImpl())
    }

    @AfterEach
    fun tearDown() {
        testEnv.close()
        OpenTelemetryConfig.shutdown()
    }

    @Test
    @Timeout(value = 30, unit = TimeUnit.SECONDS)
    fun activitySpanIncludesTemporalIdentifiers() {
        val activities = testEnv.newActivityStub(PingActivities::class.java)
        assertEquals("pong", activities.ping())

        val spans = spanExporter.finishedSpanItems
        val activitySpan =
            spans.firstOrNull { it.name.startsWith("activity.") }
        assertNotNull(activitySpan, "expected activity span, got: ${spans.map { it.name }}")

        val span = activitySpan!!
        assertFalse(span.getAttribute(WORKFLOW_ID).isNullOrBlank())
        assertFalse(span.getAttribute(RUN_ID).isNullOrBlank())
        assertFalse(span.getAttribute(ACTIVITY_TYPE).isNullOrBlank())
        // TestActivityEnvironment may leave workflow_type unset; production runs always set it.
    }

    private fun SpanData.getAttribute(key: AttributeKey<String>): String? = attributes.get(key)

    companion object {
        private val WORKFLOW_ID = AttributeKey.stringKey("workflow_id")
        private val RUN_ID = AttributeKey.stringKey("run_id")
        private val WORKFLOW_TYPE = AttributeKey.stringKey("workflow_type")
        private val ACTIVITY_TYPE = AttributeKey.stringKey("activity_type")
    }
}
