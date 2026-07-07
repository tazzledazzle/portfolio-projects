package com.portfolio.temporalobs.workflows

import io.temporal.activity.ActivityOptions
import io.temporal.workflow.Workflow
import java.time.Duration

class PingWorkflowImpl : PingWorkflow {
    private val activities: PingActivities =
        Workflow.newActivityStub(
            PingActivities::class.java,
            ActivityOptions.newBuilder()
                .setStartToCloseTimeout(Duration.ofSeconds(30))
                .build(),
        )

    override fun ping(): String = activities.ping()
}
