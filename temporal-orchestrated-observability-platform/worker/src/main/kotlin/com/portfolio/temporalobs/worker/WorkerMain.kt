package com.portfolio.temporalobs.worker

import com.portfolio.temporalobs.worker.clients.LlmClient
import com.portfolio.temporalobs.worker.telemetry.OpenTelemetryConfig
import com.portfolio.temporalobs.worker.telemetry.TemporalTracingInterceptor
import com.portfolio.temporalobs.workflows.AgentToolsWorkflowImpl
import com.portfolio.temporalobs.workflows.BatchEvalWorkflowImpl
import com.portfolio.temporalobs.workflows.PingWorkflowImpl
import com.portfolio.temporalobs.workflows.RagQaWorkflowImpl
import com.portfolio.temporalobs.workflows.TaskQueues
import com.portfolio.temporalobs.workflows.TemporalConnection
import io.temporal.worker.WorkerFactory
import io.temporal.worker.WorkerFactoryOptions
import java.util.concurrent.CountDownLatch
import java.util.concurrent.TimeUnit

fun main() {
    val otel = OpenTelemetryConfig.init(serviceVersion = "0.1.0-SNAPSHOT")
    val llmClient = LlmClient.create(otel)

    val host = TemporalConnection.temporalHost()
    val service = TemporalConnection.serviceStubs()
    val workflowClient = TemporalConnection.workflowClient(service)
    val factoryOptions =
        WorkerFactoryOptions.newBuilder()
            .setWorkerInterceptors(TemporalTracingInterceptor(otel))
            .build()
    val factory = WorkerFactory.newInstance(workflowClient, factoryOptions)

    val worker = factory.newWorker(TaskQueues.AI_WORKFLOWS)
    worker.registerWorkflowImplementationTypes(
        PingWorkflowImpl::class.java,
        RagQaWorkflowImpl::class.java,
        AgentToolsWorkflowImpl::class.java,
        BatchEvalWorkflowImpl::class.java,
    )
    worker.registerActivitiesImplementations(
        PingActivitiesImpl(),
        RagActivitiesImpl(llmClient = llmClient),
        AgentActivitiesImpl(),
        BatchActivitiesImpl(),
    )

    val shutdownLatch = CountDownLatch(1)
    Runtime.getRuntime().addShutdownHook(
        Thread {
            println("Shutting down worker (SIGTERM) ...")
            factory.shutdown()
            factory.awaitTermination(30, TimeUnit.SECONDS)
            OpenTelemetryConfig.shutdown()
            service.shutdown()
            shutdownLatch.countDown()
        },
    )

    factory.start()
    println(
        "Worker started — polling ${TaskQueues.AI_WORKFLOWS} at $host " +
            "(namespace=${TemporalConnection.namespace()}, metrics=:${System.getenv("METRICS_PORT") ?: "9464"}/metrics)",
    )

    shutdownLatch.await()
}
