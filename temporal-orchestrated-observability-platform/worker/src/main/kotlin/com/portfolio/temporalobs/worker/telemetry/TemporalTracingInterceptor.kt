package com.portfolio.temporalobs.worker.telemetry

import io.opentelemetry.api.OpenTelemetry
import io.opentelemetry.api.common.AttributeKey
import io.opentelemetry.api.trace.Span
import io.opentelemetry.api.trace.StatusCode
import io.opentelemetry.api.trace.Tracer
import io.opentelemetry.context.Scope
import io.temporal.activity.ActivityExecutionContext
import io.temporal.common.interceptors.ActivityInboundCallsInterceptor
import io.temporal.common.interceptors.ActivityInboundCallsInterceptorBase
import io.temporal.common.interceptors.WorkerInterceptor
import io.temporal.common.interceptors.WorkflowInboundCallsInterceptor
import io.temporal.common.interceptors.WorkflowInboundCallsInterceptorBase
import io.temporal.workflow.Workflow
import org.slf4j.MDC
import java.util.concurrent.TimeUnit

class TemporalTracingInterceptor(
    openTelemetry: OpenTelemetry,
) : WorkerInterceptor {
    private val tracer: Tracer = openTelemetry.getTracer("ai-temporal-worker")

    override fun interceptActivity(
        next: ActivityInboundCallsInterceptor,
    ): ActivityInboundCallsInterceptor = ActivityTracingInterceptor(next, tracer)

    override fun interceptWorkflow(
        next: WorkflowInboundCallsInterceptor,
    ): WorkflowInboundCallsInterceptor = WorkflowTracingInterceptor(next)

    private class ActivityTracingInterceptor(
        next: ActivityInboundCallsInterceptor,
        private val tracer: Tracer,
    ) : ActivityInboundCallsInterceptorBase(next) {
        private lateinit var activityContext: ActivityExecutionContext

        override fun init(context: ActivityExecutionContext) {
            super.init(context)
            activityContext = context
        }

        override fun execute(
            input: ActivityInboundCallsInterceptor.ActivityInput,
        ): ActivityInboundCallsInterceptor.ActivityOutput {
            val info = activityContext.info
            val spanName = "activity.${info.activityType}"
            val span =
                tracer.spanBuilder(spanName)
                    .setAttribute(WORKFLOW_ID, info.workflowId)
                    .setAttribute(RUN_ID, info.runId)
                    .setAttribute(WORKFLOW_TYPE, info.workflowType)
                    .setAttribute(ACTIVITY_TYPE, info.activityType)
                    .setAttribute(TASK_QUEUE, info.activityTaskQueue)
                    .startSpan()

            var scope: Scope? = null
            val startNanos = System.nanoTime()
            var status = "ok"
            try {
                scope = span.makeCurrent()
                putMdc(info, span)
                return super.execute(input)
            } catch (e: Exception) {
                status = "error"
                span.setStatus(StatusCode.ERROR, e.message ?: e.javaClass.simpleName)
                throw e
            } finally {
                val durationSeconds =
                    TimeUnit.NANOSECONDS.toMillis(System.nanoTime() - startNanos) / 1000.0
                Metrics.recordActivityDuration(
                    workflowType = info.workflowType,
                    activityType = info.activityType,
                    status = status,
                    durationSeconds = durationSeconds,
                )
                scope?.close()
                span.end()
                MDC.clear()
            }
        }

        private fun putMdc(
            info: io.temporal.activity.ActivityInfo,
            span: Span,
        ) {
            val spanContext = span.spanContext
            if (spanContext.isValid) {
                MDC.put("trace_id", spanContext.traceId)
                MDC.put("span_id", spanContext.spanId)
            }
            MDC.put("workflow_id", info.workflowId)
            MDC.put("run_id", info.runId)
            MDC.put("workflow_type", info.workflowType)
            MDC.put("activity_type", info.activityType)
        }
    }

    companion object {
        private val WORKFLOW_ID = AttributeKey.stringKey("workflow_id")
        private val RUN_ID = AttributeKey.stringKey("run_id")
        private val WORKFLOW_TYPE = AttributeKey.stringKey("workflow_type")
        private val ACTIVITY_TYPE = AttributeKey.stringKey("activity_type")
        private val TASK_QUEUE = AttributeKey.stringKey("task_queue")
    }

    private class WorkflowTracingInterceptor(
        next: WorkflowInboundCallsInterceptor,
    ) : WorkflowInboundCallsInterceptorBase(next) {
        override fun execute(
            input: WorkflowInboundCallsInterceptor.WorkflowInput,
        ): WorkflowInboundCallsInterceptor.WorkflowOutput {
            val workflowType = Workflow.getInfo().workflowType
            try {
                val output = super.execute(input)
                Metrics.incrementWorkflowCompleted(workflowType, "ok")
                return output
            } catch (e: Exception) {
                Metrics.incrementWorkflowCompleted(workflowType, "error")
                throw e
            }
        }
    }
}
