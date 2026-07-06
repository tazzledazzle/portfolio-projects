package com.portfolio.temporalobs.workflows

import com.portfolio.temporalobs.workflows.model.DatasetItem
import io.temporal.client.WorkflowOptions
import io.temporal.testing.TestWorkflowEnvironment
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Assertions.assertTrue
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.Timeout
import java.util.concurrent.TimeUnit

class BatchEvalWorkflowTest {
    private lateinit var testEnv: TestWorkflowEnvironment

    @BeforeEach
    fun setUp() {
        testEnv = TestEnvironmentFactory.newEnvironment()
        val worker = testEnv.newWorker(TaskQueues.AI_WORKFLOWS)
        worker.registerWorkflowImplementationTypes(BatchEvalWorkflowImpl::class.java)
        worker.registerActivitiesImplementations(TestBatchActivities())
        testEnv.start()
    }

    @AfterEach
    fun tearDown() {
        testEnv.close()
    }

    @Test
    @Timeout(value = 30, unit = TimeUnit.SECONDS)
    fun batchEvalWorkflowAggregatesFiveItems() {
        val workflow =
            testEnv.workflowClient.newWorkflowStub(
                BatchEvalWorkflow::class.java,
                WorkflowOptions.newBuilder().setTaskQueue(TaskQueues.AI_WORKFLOWS).build(),
            )

        val result = workflow.eval("demo-eval", 5)
        assertEquals(5, result.itemCount)
        assertTrue(result.meanScore in 0.0..1.0)
    }

    private class TestBatchActivities : BatchActivities {
        override fun loadDataset(
            datasetId: String,
            itemCount: Int,
        ): List<DatasetItem> =
            (1..itemCount).map { DatasetItem("$datasetId-$it", "prompt-$it") }

        override fun scoreItem(item: DatasetItem): Double = 0.5

        override fun aggregate(scores: List<Double>) =
            com.portfolio.temporalobs.workflows.model.BatchEvalResult(
                itemCount = scores.size,
                meanScore = scores.sum() / scores.size,
            )
    }
}
