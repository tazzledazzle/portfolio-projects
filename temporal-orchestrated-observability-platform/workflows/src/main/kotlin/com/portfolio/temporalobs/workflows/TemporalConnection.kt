package com.portfolio.temporalobs.workflows

import io.temporal.client.WorkflowClient
import io.temporal.client.WorkflowClientOptions
import io.temporal.serviceclient.WorkflowServiceStubs
import io.temporal.serviceclient.WorkflowServiceStubsOptions
import java.time.Duration

object TemporalConnection {
    fun temporalHost(): String = System.getenv("TEMPORAL_HOST") ?: "localhost:7233"

    fun namespace(): String = System.getenv("TEMPORAL_NAMESPACE") ?: "default"

    fun serviceStubs(): WorkflowServiceStubs =
        WorkflowServiceStubs.newConnectedServiceStubs(
            WorkflowServiceStubsOptions.newBuilder()
                .setTarget(temporalHost())
                .build(),
            Duration.ofSeconds(30),
        )

    fun workflowClient(stubs: WorkflowServiceStubs = serviceStubs()): WorkflowClient =
        WorkflowClient.newInstance(
            stubs,
            WorkflowClientOptions.newBuilder()
                .setNamespace(namespace())
                .setDataConverter(TemporalDataConverter.instance)
                .build(),
        )
}
