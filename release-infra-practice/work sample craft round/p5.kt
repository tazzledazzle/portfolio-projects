package p5

import java.util.ArrayDeque
import java.util.concurrent.atomic.AtomicInteger
import java.util.concurrent.locks.ReentrantLock
import kotlin.concurrent.withLock

/**
 * Change-Failure-Rate (CFR) Calculator.
 *
 * Maintains a rolling 7-day change-failure-rate per service, queryable on
 * demand. Handles deploy outcomes that are discovered after the deploy was
 * originally recorded (e.g. a rollback triggered hours later once an alert
 * fires).
 */

enum class Outcome { PENDING, SUCCESS, FAILED, ROLLED_BACK }

private fun isFailure(o: Outcome) = o == Outcome.FAILED || o == Outcome.ROLLED_BACK

const val WINDOW_SECONDS = 7.0 * 24 * 60 * 60

data class DeployEvent(
    val id: Int,
    val serviceName: String,
    val timestampSeconds: Double,
    var outcome: Outcome
)

interface CFRCalculator {
    fun recordDeploy(serviceName: String, timestampSeconds: Double, outcome: Outcome = Outcome.PENDING): Int
    fun updateOutcome(eventId: Int, outcome: Outcome)
    /** Returns null if there are no resolved deploys for the service in the window. */
    fun changeFailureRate(serviceName: String, nowSeconds: Double): Double?
}

// ---------------------------------------------------------------------------
// STEP 1 — BRUTE FORCE
// ---------------------------------------------------------------------------
// One flat list of every event, across every service, forever.
// updateOutcome scans it linearly to find the matching id. changeFailureRate
// re-scans and re-filters ALL events on every query.
// ---------------------------------------------------------------------------
class BruteForceCFRCalculator : CFRCalculator {
    private val lock = ReentrantLock()
    private val events = mutableListOf<DeployEvent>()
    private val nextId = AtomicInteger(0)

    override fun recordDeploy(serviceName: String, timestampSeconds: Double, outcome: Outcome): Int =
        lock.withLock {
            val id = nextId.incrementAndGet()
            events.add(DeployEvent(id, serviceName, timestampSeconds, outcome))
            id
        }

    override fun updateOutcome(eventId: Int, outcome: Outcome) {
        lock.withLock {
            val event = events.find { it.id == eventId } // O(n) scan to find the event
                ?: throw NoSuchElementException("no such event id: $eventId")
            event.outcome = outcome
        }
    }

    override fun changeFailureRate(serviceName: String, nowSeconds: Double): Double? = lock.withLock {
        val windowStart = nowSeconds - WINDOW_SECONDS
        var resolved = 0
        var failures = 0
        for (e in events) { // O(n) scan of ALL events, all services
            if (e.serviceName != serviceName) continue
            if (e.timestampSeconds < windowStart || e.timestampSeconds > nowSeconds) continue
            if (e.outcome == Outcome.PENDING) continue
            resolved++
            if (isFailure(e.outcome)) failures++
        }
        if (resolved == 0) null else failures.toDouble() / resolved
    }
}

// ---------------------------------------------------------------------------
// STEP 2 — OPTIMAL
// ---------------------------------------------------------------------------
// Three fixes:
//   1. Partition storage per service: a map of deques, so a query for one
//      service never scans another service's events.
//   2. Maintain running counters (resolvedCount, failureCount) per
//      service, adjusted incrementally as events enter, get corrected, and
//      age out -> O(1) query after amortized-O(1) eviction, instead of
//      O(n).
//   3. Index events by id (byId map) so updateOutcome is O(1) to locate
//      the record, not an O(n) scan.
//
// A PENDING event contributes to neither the numerator nor denominator
// until resolved — a deliberate modeling choice: counting it as an
// implicit success would bias CFR artificially low right after a burst of
// deploys, before outcomes are known.
// ---------------------------------------------------------------------------
private class ServiceCounters {
    val events = ArrayDeque<DeployEvent>()
    var resolved = 0
    var failures = 0
}

class OptimalCFRCalculator : CFRCalculator {
    private val lock = ReentrantLock()
    private val byId = HashMap<Int, DeployEvent>()
    private val perService = HashMap<String, ServiceCounters>()
    private val nextId = AtomicInteger(0)

    private fun countersFor(serviceName: String): ServiceCounters =
        perService.getOrPut(serviceName) { ServiceCounters() }

    override fun recordDeploy(serviceName: String, timestampSeconds: Double, outcome: Outcome): Int =
        lock.withLock {
            val id = nextId.incrementAndGet()
            val event = DeployEvent(id, serviceName, timestampSeconds, outcome)
            val sc = countersFor(serviceName)
            sc.events.offer(event)
            byId[id] = event
            if (outcome != Outcome.PENDING) {
                sc.resolved++
                if (isFailure(outcome)) sc.failures++
            }
            id
        }

    override fun updateOutcome(eventId: Int, outcome: Outcome) {
        lock.withLock {
            val event = byId[eventId] ?: throw NoSuchElementException("no such event id: $eventId")
            val sc = perService.getValue(event.serviceName)

            val wasResolved = event.outcome != Outcome.PENDING
            val wasFailure = isFailure(event.outcome)
            event.outcome = outcome
            val isResolved = outcome != Outcome.PENDING
            val nowFailure = isFailure(outcome)

            if (wasResolved && !isResolved) sc.resolved--
            else if (!wasResolved && isResolved) sc.resolved++
            if (wasFailure && !nowFailure) sc.failures--
            else if (!wasFailure && nowFailure) sc.failures++
            Unit
        }
    }

    private fun evictExpired(serviceName: String, nowSeconds: Double) {
        val sc = perService[serviceName] ?: return
        val windowStart = nowSeconds - WINDOW_SECONDS
        while (sc.events.isNotEmpty() && sc.events.peek().timestampSeconds < windowStart) {
            val e = sc.events.poll()
            byId.remove(e.id)
            if (e.outcome != Outcome.PENDING) {
                sc.resolved--
                if (isFailure(e.outcome)) sc.failures--
            }
        }
    }

    override fun changeFailureRate(serviceName: String, nowSeconds: Double): Double? = lock.withLock {
        evictExpired(serviceName, nowSeconds)
        val sc = perService[serviceName] ?: return@withLock null
        if (sc.resolved == 0) null else sc.failures.toDouble() / sc.resolved
    }
}
