package com.portfolio.temporalobs.workflows

import io.temporal.client.WorkflowOptions
import io.temporal.testing.TestWorkflowEnvironment
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test

class PingWorkflowTest {
    private lateinit var testEnv: TestWorkflowEnvironment

    @BeforeEach
    fun setUp() {
        testEnv = TestEnvironmentFactory.newEnvironment()
        val worker = testEnv.newWorker(TaskQueues.AI_WORKFLOWS)
        worker.registerWorkflowImplementationTypes(PingWorkflowImpl::class.java)
        worker.registerActivitiesImplementations(TestPingActivities())
        testEnv.start()
    }

    @AfterEach
    fun tearDown() {
        testEnv.close()
    }

    private class TestPingActivities : PingActivities {
        override fun ping(): String = "pong"
    }

    @Test
    fun pingWorkflowReturnsPong() {
        val workflow =
            testEnv.workflowClient.newWorkflowStub(
                PingWorkflow::class.java,
                WorkflowOptions.newBuilder()
                    .setTaskQueue(TaskQueues.AI_WORKFLOWS)
                    .build(),
            )

        assertEquals("pong", workflow.ping())
    }
}
