package com.portfolio.temporalobs.workflows

import io.temporal.client.WorkflowClientOptions
import io.temporal.testing.TestEnvironmentOptions
import io.temporal.testing.TestWorkflowEnvironment

object TestEnvironmentFactory {
    fun newEnvironment(): TestWorkflowEnvironment =
        TestWorkflowEnvironment.newInstance(
            TestEnvironmentOptions.newBuilder()
                .setWorkflowClientOptions(
                    WorkflowClientOptions.newBuilder()
                        .setDataConverter(TemporalDataConverter.instance)
                        .build(),
                )
                .build(),
        )
}
