package org.example
/*
 * P2 - Release Event Log -> DORA Metrics
 *
 * Base:  parse a raw deploy event log into typed events.
 * Ext 1: per-service deployment frequency, change failure rate, MTTR.
 * Ext 2: restrict to a rolling time window.
 * Ext 3: survive malformed lines, out-of-order events, and unresolved incidents.
 *
 * Line format: <ISO-8601 timestamp> key=value key=value ...
 *   2026-08-01T10:00:00Z service=ledger event=DEPLOY_SUCCEEDED deploy=d1
 *   2026-08-01T10:40:00Z service=ledger event=INCIDENT_OPENED incident=i1 deploy=d1
 *   2026-08-01T11:10:00Z service=ledger event=INCIDENT_RESOLVED incident=i1
 */

import java.time.Duration
import java.time.Instant
import java.time.format.DateTimeParseException

enum class EventType { DEPLOY_SUCCEEDED, DEPLOY_FAILED, INCIDENT_OPENED, INCIDENT_RESOLVED }

data class ReleaseEvent(
    val at: Instant,
    val service: String,
    val type: EventType,
    val deployId: String? = null,
    val incidentId: String? = null
)

data class ParseReport(val events: List<ReleaseEvent>, val rejected: List<Pair<Int, String>>)

data class ServiceMetrics(
    val service: String,
    val successfulDeploys: Int,
    val failedDeploys: Int,
    val deploysPerDay: Double,
    val changeFailureRate: Double,
    val meanTimeToRestore: Duration?,
    val unresolvedIncidents: Int
) {
    override fun toString(): String = buildString {
        append(service.padEnd(12))
        append("deploys=%-4d".format(successfulDeploys))
        append("freq=%-8s".format("%.2f/d".format(deploysPerDay)))
        append("cfr=%-8s".format("%.0f%%".format(changeFailureRate * 100)))
        append("mttr=%-10s".format(meanTimeToRestore?.let { "${it.toMinutes()}m" } ?: "n/a"))
        append("open=$unresolvedIncidents")
    }
}

object ReleaseLog {

    /** Never throws. Bad lines go to `rejected` with their 1-based line number. */
    fun parse(lines: List<String>): ParseReport {
        val events = mutableListOf<ReleaseEvent>()
        val rejected = mutableListOf<Pair<Int, String>>()

        lines.forEachIndexed { idx, raw ->
            val lineNo = idx + 1
            val line = raw.trim()
            if (line.isEmpty() || line.startsWith("#")) return@forEachIndexed

            val tokens = line.split(Regex("\\s+"))
            val at = try {
                Instant.parse(tokens.first())
            } catch (e: DateTimeParseException) {
                rejected += lineNo to "bad timestamp: ${tokens.first()}"
                return@forEachIndexed
            }

            val fields = mutableMapOf<String, String>()
            for (token in tokens.drop(1)) {
                val eq = token.indexOf('=')
                if (eq <= 0) {
                    rejected += lineNo to "malformed field: $token"
                    return@forEachIndexed
                }
                fields[token.substring(0, eq)] = token.substring(eq + 1)
            }

            val service = fields["service"]
            val typeRaw = fields["event"]
            if (service == null || typeRaw == null) {
                rejected += lineNo to "missing service or event"
                return@forEachIndexed
            }
            val type = EventType.entries.firstOrNull { it.name == typeRaw }
            if (type == null) {
                rejected += lineNo to "unknown event type: $typeRaw"
                return@forEachIndexed
            }
            events += ReleaseEvent(at, service, type, fields["deploy"], fields["incident"])
        }
        // Producers batch and retry, so arrival order is not event order. Sort once here.
        return ParseReport(events.sortedBy { it.at }, rejected)
    }

    /**
     * `window` bounds which events count. Deploy frequency is normalised over the
     * observed span (or `window` when supplied) so short logs are not flattered.
     */
    fun metrics(
        events: List<ReleaseEvent>,
        from: Instant? = null,
        to: Instant? = null
    ): List<ServiceMetrics> {
        val scoped = events.filter { (from == null || !it.at.isBefore(from)) && (to == null || !it.at.isAfter(to)) }
        if (scoped.isEmpty()) return emptyList()

        val spanStart = from ?: scoped.first().at
        val spanEnd = to ?: scoped.last().at
        val spanDays = maxOf(Duration.between(spanStart, spanEnd).toMillis() / 86_400_000.0, 1.0 / 24)

        return scoped.groupBy { it.service }.map { (service, evts) ->
            val succeeded = evts.count { it.type == EventType.DEPLOY_SUCCEEDED }
            val failed = evts.count { it.type == EventType.DEPLOY_FAILED }

            val opened = evts.filter { it.type == EventType.INCIDENT_OPENED }
            val resolvedAt = evts.filter { it.type == EventType.INCIDENT_RESOLVED }
                .associate { it.incidentId to it.at }

            // A deploy is "failed change" if any incident names it. Distinct, so two
            // incidents from one bad deploy do not double-count.
            val blamedDeploys = opened.mapNotNull { it.deployId }.toSet()
            val cfr = if (succeeded == 0) 0.0 else blamedDeploys.size.toDouble() / succeeded

            val restoreTimes = opened.mapNotNull { inc ->
                val end = resolvedAt[inc.incidentId] ?: return@mapNotNull null
                Duration.between(inc.at, end)
            }
            val mttr = if (restoreTimes.isEmpty()) null
            else Duration.ofSeconds(restoreTimes.sumOf { it.seconds } / restoreTimes.size)

            ServiceMetrics(
                service = service,
                successfulDeploys = succeeded,
                failedDeploys = failed,
                deploysPerDay = succeeded / spanDays,
                changeFailureRate = cfr,
                meanTimeToRestore = mttr,
                unresolvedIncidents = opened.size - restoreTimes.size
            )
        }.sortedBy { it.service }
    }
}

// ---------------------------------------------------------------- demo / tests

private fun check(label: String, condition: Boolean) {
    if (!condition) throw AssertionError("FAILED: $label")
    println("  ok  $label")
}

fun main() {
    println("P2 - Release Event Log -> DORA Metrics")

    val log = """
        # deploy pipeline audit log
        2026-08-01T09:00:00Z service=ledger  event=DEPLOY_SUCCEEDED deploy=d1
        2026-08-01T10:40:00Z service=ledger  event=INCIDENT_OPENED   incident=i1 deploy=d1
        2026-08-01T11:10:00Z service=ledger  event=INCIDENT_RESOLVED incident=i1
        2026-08-01T09:30:00Z service=web     event=DEPLOY_SUCCEEDED deploy=d2
        2026-08-02T09:00:00Z service=ledger  event=DEPLOY_SUCCEEDED deploy=d3
        2026-08-02T09:05:00Z service=ledger  event=DEPLOY_FAILED    deploy=d4
        2026-08-02T12:00:00Z service=web     event=DEPLOY_SUCCEEDED deploy=d5
        2026-08-02T13:00:00Z service=web     event=INCIDENT_OPENED   incident=i2 deploy=d5
        not-a-timestamp service=web event=DEPLOY_SUCCEEDED
        2026-08-02T14:00:00Z service=web     event=EXPLODED
        2026-08-03T09:00:00Z service=ledger  event=DEPLOY_SUCCEEDED deploy=d6
    """.trimIndent().lines()

    val report = ReleaseLog.parse(log)
    println("parsed ${report.events.size} events, rejected ${report.rejected.size}")
    report.rejected.forEach { (line, why) -> println("   line $line: $why") }

    check("9 good events parsed", report.events.size == 9)
    check("2 lines rejected", report.rejected.size == 2)
    check("events sorted by time", report.events.zipWithNext().all { !it.first.at.isAfter(it.second.at) })

    val metrics = ReleaseLog.metrics(report.events)
    metrics.forEach { println("   $it") }

    val ledger = metrics.first { it.service == "ledger" }
    val web = metrics.first { it.service == "web" }

    check("ledger had 3 successful deploys", ledger.successfulDeploys == 3)
    check("ledger had 1 failed deploy", ledger.failedDeploys == 1)
    check("ledger CFR is 1/3", "%.2f".format(ledger.changeFailureRate) == "0.33")
    check("ledger MTTR is 30 minutes", ledger.meanTimeToRestore == Duration.ofMinutes(30))
    check("web has an unresolved incident", web.unresolvedIncidents == 1)
    check("web MTTR is null with nothing resolved", web.meanTimeToRestore == null)

    // Ext 2 - windowing
    val dayTwo = ReleaseLog.metrics(
        report.events,
        from = Instant.parse("2026-08-02T00:00:00Z"),
        to = Instant.parse("2026-08-02T23:59:59Z")
    )
    check("window excludes day-1 deploys",
        dayTwo.first { it.service == "ledger" }.successfulDeploys == 1)
    check("empty window yields no rows",
        ReleaseLog.metrics(report.events,
            from = Instant.parse("2027-01-01T00:00:00Z"),
            to = Instant.parse("2027-01-02T00:00:00Z")).isEmpty())

    check("empty log is safe", ReleaseLog.parse(emptyList()).events.isEmpty())

    println("P2 passed.\n")
}
