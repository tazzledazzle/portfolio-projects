package com.portfolio.temporalobs.workflows

import com.portfolio.temporalobs.workflows.model.AgentToolsResult
import io.temporal.workflow.WorkflowInterface
import io.temporal.workflow.WorkflowMethod

@WorkflowInterface
interface AgentToolsWorkflow {
    @WorkflowMethod
    fun run(goal: String): AgentToolsResult
}
