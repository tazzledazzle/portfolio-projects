package com.portfolio.temporalobs.workflows

import com.portfolio.temporalobs.workflows.model.RagQaResult
import io.temporal.workflow.WorkflowInterface
import io.temporal.workflow.WorkflowMethod

@WorkflowInterface
interface RagQaWorkflow {
    @WorkflowMethod
    fun ask(question: String): RagQaResult
}
