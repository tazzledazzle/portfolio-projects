package com.portfolio.temporalobs.worker.telemetry

import io.opentelemetry.api.OpenTelemetry
import io.opentelemetry.api.common.Attributes
import io.opentelemetry.api.metrics.DoubleHistogram
import io.opentelemetry.api.metrics.LongCounter
import io.opentelemetry.api.metrics.Meter

object Metrics {
    private const val METER_NAME = "ai-temporal-worker"

    private var activityDuration: DoubleHistogram? = null
    private var workflowCompleted: LongCounter? = null
    private var llmRequestDuration: DoubleHistogram? = null

    fun bind(openTelemetry: OpenTelemetry) {
        val meter: Meter = openTelemetry.getMeter(METER_NAME)
        activityDuration =
            meter.histogramBuilder("activity.duration")
                .setDescription("Activity execution duration")
                .setUnit("s")
                .build()
        workflowCompleted =
            meter.counterBuilder("workflow.completed")
                .setDescription("Workflow terminal outcomes")
                .build()
        llmRequestDuration =
            meter.histogramBuilder("llm.request.duration")
                .setDescription("LLM HTTP client request duration")
                .setUnit("s")
                .build()
    }

    fun unbind() {
        activityDuration = null
        workflowCompleted = null
        llmRequestDuration = null
    }

    fun recordActivityDuration(
        workflowType: String,
        activityType: String,
        status: String,
        durationSeconds: Double,
    ) {
        activityDuration?.record(
            durationSeconds,
            Attributes.builder()
                .put("workflow_type", workflowType)
                .put("activity_type", activityType)
                .put("status", status)
                .build(),
        )
    }

    fun incrementWorkflowCompleted(
        workflowType: String,
        status: String,
    ) {
        workflowCompleted?.add(
            1,
            Attributes.builder()
                .put("workflow_type", workflowType)
                .put("status", status)
                .build(),
        )
    }

    fun recordLlmDuration(
        status: String,
        durationSeconds: Double,
    ) {
        llmRequestDuration?.record(
            durationSeconds,
            Attributes.builder()
                .put("status", status)
                .build(),
        )
    }
}
