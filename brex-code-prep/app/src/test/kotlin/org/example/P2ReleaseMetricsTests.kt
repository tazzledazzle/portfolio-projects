package org.example

import org.junit.jupiter.api.Nested
import org.junit.jupiter.api.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

class P2ReleaseMetricsTests {
    @Nested
    inner class Parsing {
        @Test
        fun `valid lines parse into typed events`() {
            val lines = listOf(
                "2026-08-01T09:00:00Z service=ledger event=DEPLOY_SUCCEEDED deploy=d1",
                "2026-08-01T09:30:00Z service=ledger event=INCIDENT_OPENED incident=i1 deploy=d1",
                "2026-08-01T10:00:00Z service=ledger event=INCIDENT_RESOLVED incident=i1"
            )
            val report = ReleaseLog.parse(lines)
            assertEquals(3, report.events.size)
            assertTrue(report.rejected.isEmpty())
        }

        @Test
        fun `bad timestamp line is rejected not thrown`() {
            val lines = listOf(
                "not-a-timestamp service=ledger event=DEPLOY_SUCCEEDED deploy=d1"
            )
            val report = ReleaseLog.parse(lines)
            assertTrue(report.events.isEmpty())
            assertEquals(1, report.rejected.size)
        }

        @Test
        fun `unknown event type is rejected`() {
            val lines = listOf(
                "2026-08-01T09:00:00Z service=ledger event=EXPLODED"
            )
            val report = ReleaseLog.parse(lines)
            assertTrue(report.events.isEmpty())
            assertEquals(1, report.rejected.size)
        }

        @Test
        fun `comment lines and blank lines are silently skipped`(){
            val lines = listOf(
                "# this is a comment",
                "",
                "2026-08-01T09:00:00Z service=ledger event=DEPLOY_SUCCEEDED deploy=d1"
            )
            val report = ReleaseLog.parse(lines)
            assertEquals(1, report.events.size)
            assertTrue(report.rejected.isEmpty())
        }

        @Test
        fun `out-of-order events are sorted by timestamp`() {
            val lines = listOf(
                "2026-08-01T10:00:00Z service=ledger event=DEPLOY_SUCCEEDED deploy=d2",
                "2026-08-01T09:00:00Z service=ledger event=DEPLOY_SUCCEEDED deploy=d1"
            )
            val report = ReleaseLog.parse(lines)
            assertEquals("d1", report.events[0].deployId)
            assertEquals("d2", report.events[1].deployId)
        }

        @Test
        fun `empty input returns empty report`(){
            val report = ReleaseLog.parse(emptyList())
            assertTrue(report.events.isEmpty())
            assertTrue(report.rejected.isEmpty())
        }
    }
}
