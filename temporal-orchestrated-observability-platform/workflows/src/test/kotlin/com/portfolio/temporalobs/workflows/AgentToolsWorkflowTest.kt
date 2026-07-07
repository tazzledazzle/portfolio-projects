package com.portfolio.temporalobs.workflows

import com.portfolio.temporalobs.workflows.model.AgentPlan
import io.temporal.client.WorkflowOptions
import io.temporal.testing.TestWorkflowEnvironment
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Assertions.assertTrue
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test

class AgentToolsWorkflowTest {
    private lateinit var testEnv: TestWorkflowEnvironment

    @BeforeEach
    fun setUp() {
        testEnv = TestEnvironmentFactory.newEnvironment()
        val worker = testEnv.newWorker(TaskQueues.AI_WORKFLOWS)
        worker.registerWorkflowImplementationTypes(AgentToolsWorkflowImpl::class.java)
        worker.registerActivitiesImplementations(TestAgentActivities())
        testEnv.start()
    }

    @AfterEach
    fun tearDown() {
        testEnv.close()
    }

    @Test
    fun agentToolsWorkflowCompletes() {
        val workflow =
            testEnv.workflowClient.newWorkflowStub(
                AgentToolsWorkflow::class.java,
                WorkflowOptions.newBuilder().setTaskQueue(TaskQueues.AI_WORKFLOWS).build(),
            )

        val result = workflow.run("summarize architecture")
        assertTrue(result.summary.contains("tool-result"))
        assertEquals(1, result.toolCalls)
    }

    private class TestAgentActivities : AgentActivities {
        override fun planStep(goal: String): AgentPlan = AgentPlan("search", goal)

        override fun callTool(
            toolName: String,
            arguments: String,
        ): String = "tool-result:$toolName"

        override fun synthesize(
            goal: String,
            toolResult: String,
        ): String = "Synthesis for '$goal' using $toolResult"
    }
}
