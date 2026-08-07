package org.example

import org.junit.jupiter.api.Nested
import kotlin.test.Test
import kotlin.test.assertEquals

class P1SchedulerTests {
    @Nested // @Nested lets us group by scenario cleanly
    inner class HappyPath {

        @Test fun `linear chain deploys in order`() {
            val jobs = listOf(
                Job("a"),
                Job("b", dependsOn = listOf("a")),
                Job("c", dependsOn = listOf("b")),
            )
            val result = DeployScheduler(jobs).plan() as ScheduleResult.Ok // if result type wrong ClassCastException is info more than null-check
            assertEquals(listOf(listOf("a"), listOf("b"), listOf("c")), result.waves)
        }

        @Test
        fun `independent jobs land in the same wave`() {
            val jobs = listOf(Job("x"), Job("y"), Job("z"))
            val result = DeployScheduler(jobs).plan() as ScheduleResult.Ok
            assertEquals(1, result.waves.size)
            assertEquals(listOf("x", "y", "z"), result.waves[0])
        }

    }
}
