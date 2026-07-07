package com.portfolio.temporalobs.workflows

import io.temporal.workflow.WorkflowInterface
import io.temporal.workflow.WorkflowMethod

@WorkflowInterface
interface PingWorkflow {
    @WorkflowMethod
    fun ping(): String
}
