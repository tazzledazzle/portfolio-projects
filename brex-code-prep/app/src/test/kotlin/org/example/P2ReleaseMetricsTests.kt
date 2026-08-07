package org.example

import org.junit.jupiter.api.Nested
import org.junit.jupiter.api.Test
import java.time.Instant
import kotlin.test.assertEquals
import kotlin.test.assertNull
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

    @Nested
    inner class Metrics {

        private val events = ReleaseLog.parse(listOf(
            "2026-08-01T09:00:00Z service=ledger event=DEPLOY_SUCCEEDED deploy=d1",
            "2026-08-01T10:40:00Z service=ledger event=INCIDENT_OPENED  incident=i1 deploy=d1",
            "2026-08-01T11:10:00Z service=ledger event=INCIDENT_RESOLVED incident=i1",
            "2026-08-02T09:00:00Z service=ledger event=DEPLOY_SUCCEEDED deploy=d3",
            "2026-08-02T09:05:00Z service=ledger event=DEPLOY_FAILED    deploy=d4",
            "2026-08-02T12:00:00Z service=web    event=DEPLOY_SUCCEEDED deploy=d5",
            "2026-08-02T13:00:00Z service=web    event=INCIDENT_OPENED  incident=i2 deploy=d5"
        )).events

        @Test
        fun `successful and failed deploy counts are correct`() {
            val ledger = ReleaseLog.metrics(events).first { it.service == "ledger" }
            assertEquals(2, ledger.successfulDeploys)
            assertEquals(1, ledger.failedDeploys)
        }

        @Test
        fun `CFR counts blamed deploys not incident count`() {
            // one incident blaming d1 → CFR = 1 blamed / 2 succeeded = 0.50
            val ledger = ReleaseLog.metrics(events).first { it.service == "ledger" }
            assertEquals(0.5, ledger.changeFailureRate, 0.01)
        }

        @Test
        fun `MTTR is average restore time for resolved incidents`() {
            // i1 opened 10:40, resolved 11:10 = 30 min
            val ledger = ReleaseLog.metrics(events).first { it.service == "ledger" }
//            val expected = 30.toDuration(DurationUnit.MINUTES)
            val expected = java.time.Duration.ofMinutes(30)
            assertEquals<java.time.Duration>(expected = expected, actual = ledger.meanTimeToRestore!!)
        }

        @Test
        fun `unresolved incident does not contribute to MTTR`() {
            val web = ReleaseLog.metrics(events).first { it.service == "web" }
            assertNull(web.meanTimeToRestore)
            assertEquals(1, web.unresolvedIncidents)
        }

        @Test
        fun `two incidents on one deploy count as one failure for CFR`() {
            val extraEvents = ReleaseLog.parse(listOf(
                "2026-08-01T09:00:00Z service=svc event=DEPLOY_SUCCEEDED deploy=bad",
                "2026-08-01T10:00:00Z service=svc event=INCIDENT_OPENED  incident=x deploy=bad",
                "2026-08-01T11:00:00Z service=svc event=INCIDENT_OPENED  incident=y deploy=bad",
                "2026-08-01T10:30:00Z service=svc event=INCIDENT_RESOLVED incident=x",
                "2026-08-01T11:30:00Z service=svc event=INCIDENT_RESOLVED incident=y"
            )).events
            val m = ReleaseLog.metrics(extraEvents).first()
            // 1 blamed deploy out of 1 succeeded = 1.0, not 2.0
            assertEquals(1.0, m.changeFailureRate, 0.001)
        }

        @Test
        fun `service with no incidents has zero CFR and null MTTR`() {
            val clean = ReleaseLog.parse(listOf(
                "2026-08-01T09:00:00Z service=clean event=DEPLOY_SUCCEEDED deploy=d1"
            )).events
            val m = ReleaseLog.metrics(clean).first()
            assertEquals(0.0, m.changeFailureRate)
            assertNull(m.meanTimeToRestore)
        }
    }

    @Nested
    inner class Windowing {

        private val allEvents = ReleaseLog.parse(listOf(
            "2026-08-01T09:00:00Z service=ledger event=DEPLOY_SUCCEEDED deploy=d1",
            "2026-08-02T09:00:00Z service=ledger event=DEPLOY_SUCCEEDED deploy=d2",
            "2026-08-03T09:00:00Z service=ledger event=DEPLOY_SUCCEEDED deploy=d3"
        )).events

        @Test
        fun `from filter excludes events before the window`() {
            val result = ReleaseLog.metrics(
                allEvents,
                from = Instant.parse("2026-08-02T00:00:00Z")
            )
            assertEquals(2, result.first().successfulDeploys)
        }

        @Test
        fun `to filter excludes events after the window`() {
            val result = ReleaseLog.metrics(
                allEvents,
                to = Instant.parse("2026-08-01T23:59:59Z")
            )
            assertEquals(1, result.first().successfulDeploys)
        }

        @Test
        fun `window with no matching events returns empty list`() {
            val result = ReleaseLog.metrics(
                allEvents,
                from = Instant.parse("2030-01-01T00:00:00Z"),
                to   = Instant.parse("2030-01-02T00:00:00Z")
            )
            assertTrue(result.isEmpty())
        }
    }
}
