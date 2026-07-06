package com.portfolio.temporalobs.workflows

import com.portfolio.temporalobs.workflows.model.AgentToolsResult
import io.temporal.activity.ActivityOptions
import io.temporal.common.RetryOptions
import io.temporal.workflow.Workflow
import java.time.Duration

class AgentToolsWorkflowImpl : AgentToolsWorkflow {
    private val activities: AgentActivities =
        Workflow.newActivityStub(
            AgentActivities::class.java,
            ActivityOptions.newBuilder()
                .setStartToCloseTimeout(Duration.ofMinutes(2))
                .build(),
        )

    private val toolActivities: AgentActivities =
        Workflow.newActivityStub(
            AgentActivities::class.java,
            ActivityOptions.newBuilder()
                .setStartToCloseTimeout(Duration.ofMinutes(2))
                .setRetryOptions(
                    RetryOptions.newBuilder()
                        .setMaximumAttempts(3)
                        .setInitialInterval(Duration.ofSeconds(1))
                        .setMaximumInterval(Duration.ofSeconds(5))
                        .setBackoffCoefficient(1.5)
                        .build(),
                )
                .build(),
        )

    override fun run(goal: String): AgentToolsResult {
        val plan = activities.planStep(goal)
        val toolResult = toolActivities.callTool(plan.toolName, plan.arguments)
        val summary = activities.synthesize(goal, toolResult)
        return AgentToolsResult(summary = summary, toolCalls = 1)
    }
}
