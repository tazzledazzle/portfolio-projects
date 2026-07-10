package com.patterns.saga

import io.temporal.client.WorkflowClient
import io.temporal.client.WorkflowOptions
import io.temporal.serviceclient.WorkflowServiceStubs
import io.temporal.worker.Worker
import io.temporal.worker.WorkerFactory
import org.slf4j.LoggerFactory
import java.time.Duration
import java.util.UUID

const val TASK_QUEUE = "order-saga-queue"

private val log = LoggerFactory.getLogger("SagaWorker")

/**
 * Temporal worker — registers workflow and activity implementations.
 *
 * Prerequisites:
 *   1. Temporal server running locally: docker run -p 7233:7233 temporalio/auto-setup:latest
 *   2. Temporal Web UI available at http://localhost:8233
 *
 * Run: gradle run
 * The worker will process workflows submitted to the "order-saga-queue" task queue.
 */
fun main() {
    val stubs = WorkflowServiceStubs.newLocalServiceStubs()
    val client = WorkflowClient.newInstance(stubs)
    val factory = WorkerFactory.newInstance(client)

    val worker: Worker = factory.newWorker(TASK_QUEUE)

    // Register the workflow implementation
    worker.registerWorkflowImplementationTypes(OrderSagaWorkflowImpl::class.java)

    // Register the activity implementation
    worker.registerActivitiesImplementations(OrderActivitiesImpl())

    factory.start()
    log.info("Worker started. Listening on task queue: {}", TASK_QUEUE)
    log.info("Open http://localhost:8233 to see workflow executions in the Temporal Web UI")

    // Submit a demo workflow to show it working
    submitDemoWorkflow(client)

    // Keep the worker running
    Runtime.getRuntime().addShutdownHook(Thread {
        log.info("Shutting down worker…")
        factory.shutdown()
    })

    // Block main thread — worker runs on background threads
    Thread.currentThread().join()
}

private fun submitDemoWorkflow(client: WorkflowClient) {
    val orderId = "order-${UUID.randomUUID()}"

    val workflow = client.newWorkflowStub(
        OrderSagaWorkflow::class.java,
        WorkflowOptions.newBuilder()
            .setTaskQueue(TASK_QUEUE)
            .setWorkflowId("order-saga-$orderId")
            .setWorkflowExecutionTimeout(Duration.ofMinutes(5))
            .build(),
    )

    val order = OrderRequest(
        orderId = orderId,
        customerId = "cust-demo",
        items = listOf(
            OrderItem("SKU-WIDGET", quantity = 2, unitPriceCents = 999L),
            OrderItem("SKU-GADGET", quantity = 1, unitPriceCents = 2999L),
        ),
        totalCents = 4997L,
    )

    log.info("Submitting saga for orderId={}", orderId)

    // Fire-and-forget — the worker processes it asynchronously
    WorkflowClient.start(workflow::execute, order)

    log.info("Workflow submitted. workflowId=order-saga-{}", orderId)
    log.info("Check Temporal UI at http://localhost:8233 to trace execution")
}
