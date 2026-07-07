package com.portfolio.temporalobs.workflows

import com.portfolio.temporalobs.workflows.model.BatchEvalResult
import io.temporal.activity.ActivityOptions
import io.temporal.workflow.Async
import io.temporal.workflow.Workflow
import java.time.Duration

class BatchEvalWorkflowImpl : BatchEvalWorkflow {
    private val activities: BatchActivities =
        Workflow.newActivityStub(
            BatchActivities::class.java,
            ActivityOptions.newBuilder()
                .setStartToCloseTimeout(Duration.ofMinutes(5))
                .build(),
        )

    override fun eval(
        datasetId: String,
        itemCount: Int,
    ): BatchEvalResult {
        val items = activities.loadDataset(datasetId, itemCount)
        val scores =
            items.map { item ->
                Async.function(activities::scoreItem, item).get()
            }
        return activities.aggregate(scores)
    }
}
