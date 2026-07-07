package com.portfolio.temporalobs.workflows

import com.portfolio.temporalobs.workflows.model.BatchEvalResult
import io.temporal.workflow.WorkflowInterface
import io.temporal.workflow.WorkflowMethod

@WorkflowInterface
interface BatchEvalWorkflow {
    @WorkflowMethod
    fun eval(
        datasetId: String,
        itemCount: Int,
    ): BatchEvalResult
}
