package org.example

import org.junit.jupiter.api.Nested
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

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

        @Test
        fun `diamond dependency resolves correctly`() {
            // proto -> common-lib -> ledger \
            //       -> card-auth     -> web
            val jobs = listOf(
                Job("proto"),
                Job("common-lib", dependsOn = listOf("proto")),
                Job("ledger", dependsOn = listOf("common-lib")),
                Job("card-auth", dependsOn = listOf("common-lib")),
                Job("web", dependsOn = listOf("ledger", "card-auth")),
            )
            val result = DeployScheduler(jobs).plan() as ScheduleResult.Ok
            val order = result.waves.flatten()
            assertTrue(order.indexOf("proto") < order.indexOf("common-lib"))
            assertTrue(order.indexOf("common-lib") < order.indexOf("ledger"))
            assertTrue(order.indexOf("ledger") < order.indexOf("web"))
            assertTrue(order.indexOf("card-auth") < order.indexOf("web"))
            assertEquals(listOf("web"), result.waves.last())
        }

        @Test
        fun `empty graph returns no waves`(){
            val result = DeployScheduler(emptyList()).plan() as ScheduleResult.Ok
            assertTrue(result.waves.isEmpty())
        }

        @Test
        fun `single node with deps is wave 0`(){
            val result = DeployScheduler(listOf(Job("solo"))).plan() as ScheduleResult.Ok
            assertEquals(listOf(listOf("solo")), result.waves)
        }
    }
}
