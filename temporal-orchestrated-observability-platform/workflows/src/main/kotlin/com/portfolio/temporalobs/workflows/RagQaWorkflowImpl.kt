package com.portfolio.temporalobs.workflows

import com.portfolio.temporalobs.workflows.model.RagQaResult
import io.temporal.activity.ActivityOptions
import io.temporal.workflow.Workflow
import java.time.Duration

class RagQaWorkflowImpl : RagQaWorkflow {
    private val activities: RagActivities =
        Workflow.newActivityStub(
            RagActivities::class.java,
            ActivityOptions.newBuilder()
                .setStartToCloseTimeout(Duration.ofMinutes(2))
                .build(),
        )

    override fun ask(question: String): RagQaResult {
        val embedding = activities.embedQuery(question)
        val chunks = activities.vectorSearch(embedding)
        val completion = activities.llmComplete(question, chunks)
        val parts = completion.split("||", limit = 2)
        val answer = parts[0]
        val citation = parts.getOrElse(1) { "unknown" }
        return RagQaResult(answer = answer, citation = citation)
    }
}
