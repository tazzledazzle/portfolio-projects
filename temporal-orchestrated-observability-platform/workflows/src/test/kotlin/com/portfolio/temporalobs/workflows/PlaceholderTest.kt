package com.portfolio.temporalobs.workflows

import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Test

class PlaceholderTest {
    @Test
    fun taskQueueIsDefined() {
        assertEquals("ai-workflows", TaskQueues.AI_WORKFLOWS)
    }
}
