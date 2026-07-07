package com.portfolio.temporalobs.worker

import com.portfolio.temporalobs.workflows.AgentActivities
import com.portfolio.temporalobs.workflows.model.AgentPlan
import io.temporal.activity.Activity
import org.slf4j.LoggerFactory

class AgentActivitiesImpl : AgentActivities {
    override fun planStep(goal: String): AgentPlan {
        logger.info("planStep goal={}", goal)
        return AgentPlan(toolName = "search_docs", arguments = goal)
    }

    override fun callTool(
        toolName: String,
        arguments: String,
    ): String {
        val attempt = Activity.getExecutionContext().info.attempt
        logger.info("callTool tool={} attempt={}", toolName, attempt)

        if (simulateToolFailure() && attempt == 1) {
            throw IllegalStateException("Simulated tool failure (attempt 1)")
        }

        return "tool-result:$toolName:${arguments.take(80)}"
    }

    override fun synthesize(
        goal: String,
        toolResult: String,
    ): String {
        logger.info("synthesize goal={}", goal)
        return "Synthesis for '$goal' using $toolResult"
    }

    companion object {
        private val logger = LoggerFactory.getLogger(AgentActivitiesImpl::class.java)

        private fun simulateToolFailure(): Boolean =
            System.getenv("SIMULATE_TOOL_FAILURE")?.equals("true", ignoreCase = true) == true
    }
}
