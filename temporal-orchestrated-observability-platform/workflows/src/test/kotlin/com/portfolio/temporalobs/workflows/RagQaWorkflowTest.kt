package com.portfolio.temporalobs.workflows

import io.temporal.client.WorkflowOptions
import io.temporal.testing.TestWorkflowEnvironment
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Assertions.assertTrue
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test

class RagQaWorkflowTest {
    private lateinit var testEnv: TestWorkflowEnvironment

    @BeforeEach
    fun setUp() {
        testEnv = TestEnvironmentFactory.newEnvironment()
        val worker = testEnv.newWorker(TaskQueues.AI_WORKFLOWS)
        worker.registerWorkflowImplementationTypes(RagQaWorkflowImpl::class.java)
        worker.registerActivitiesImplementations(TestRagActivities())
        testEnv.start()
    }

    @AfterEach
    fun tearDown() {
        testEnv.close()
    }

    @Test
    fun ragQaWorkflowReturnsAnswerWithCitation() {
        val workflow =
            testEnv.workflowClient.newWorkflowStub(
                RagQaWorkflow::class.java,
                WorkflowOptions.newBuilder().setTaskQueue(TaskQueues.AI_WORKFLOWS).build(),
            )

        val result = workflow.ask("How does observability work?")
        assertTrue(result.answer.isNotBlank())
        assertEquals("fixtures/chunks.json#chunk-0", result.citation)
    }

    private class TestRagActivities : RagActivities {
        override fun embedQuery(question: String): String = "emb-test"

        override fun vectorSearch(embedding: String): String =
            "docs/ARCHITECTURE.md#chunk-0: observability text"

        override fun llmComplete(
            question: String,
            chunks: String,
        ): String = "Answer for $question||fixtures/chunks.json#chunk-0"
    }
}
