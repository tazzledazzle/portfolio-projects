package com.portfolio.temporalobs.workflows

import com.portfolio.temporalobs.workflows.model.AgentPlan
import io.temporal.activity.ActivityInterface
import io.temporal.activity.ActivityMethod

@ActivityInterface
interface AgentActivities {
    @ActivityMethod
    fun planStep(goal: String): AgentPlan

    @ActivityMethod
    fun callTool(
        toolName: String,
        arguments: String,
    ): String

    @ActivityMethod
    fun synthesize(
        goal: String,
        toolResult: String,
    ): String
}
