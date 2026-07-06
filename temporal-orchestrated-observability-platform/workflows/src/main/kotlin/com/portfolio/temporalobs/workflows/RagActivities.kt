package com.portfolio.temporalobs.workflows

import io.temporal.activity.ActivityInterface
import io.temporal.activity.ActivityMethod

@ActivityInterface
interface RagActivities {
    @ActivityMethod
    fun embedQuery(question: String): String

    @ActivityMethod
    fun vectorSearch(embedding: String): String

    /** Returns `answer text` and `citation` separated by the delimiter `||`. */
    @ActivityMethod
    fun llmComplete(
        question: String,
        chunks: String,
    ): String
}
