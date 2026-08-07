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
    }
}
