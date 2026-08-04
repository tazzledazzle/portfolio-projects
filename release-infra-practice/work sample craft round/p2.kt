package p2

import java.util.ArrayDeque
import java.util.concurrent.ConcurrentHashMap
import java.util.concurrent.CopyOnWriteArrayList
import java.util.concurrent.locks.ReentrantLock
import kotlin.concurrent.withLock

/**
 * Rate-Limited Rollback Endpoint.
 *
 * At most 3 rollbacks per service per 10-minute SLIDING window. Returns
 * rejection + retryAfterSeconds when exceeded. Every ACCEPTED rollback is
 * written to an append-only audit log.
 */

const val WINDOW_MILLIS = 10 * 60 * 1000L
const val MAX_ROLLBACKS_PER_WINDOW = 3

data class RollbackResult(
    val accepted: Boolean,
    val retryAfterSeconds: Double? = null,
    val auditId: String? = null
)

data class AuditEntry(val id: String, val serviceName: String, val requestedAtMillis: Long)

/**
 * Durable, append-only. Production backing would be a file opened
 * O_APPEND + fsync, or an INSERT-only DB table. The interesting design
 * decisions (what gets logged, and WHEN relative to the rate-limit
 * decision) are independent of storage backend, so a thread-safe
 * in-memory list is enough to demonstrate them.
 */
class AppendOnlyAuditLog {
    private val entries = CopyOnWriteArrayList<AuditEntry>()
    private var counter = 0L
    private val counterLock = ReentrantLock()

    fun append(serviceName: String, requestedAtMillis: Long): String {
        val id = counterLock.withLock { (++counter).toString() } + "-$serviceName"
        entries.add(AuditEntry(id, serviceName, requestedAtMillis))
        return id
    }

    fun allEntries(): List<AuditEntry> = entries.toList()
}

// ---------------------------------------------------------------------------
// STEP 1 — BRUTE FORCE
// ---------------------------------------------------------------------------
// A plain MutableList<Long> (timestamps) per service. Every call rebuilds
// the "still inside the window" list via a fresh filter (O(n)), checks the
// size, and appends. "Read, decide, write" are three separate,
// unsynchronized steps -> two concurrent requests can both read size 2,
// both decide "allowed", and both append, letting 4 rollbacks through a
// limit of 3.
// ---------------------------------------------------------------------------
class BruteForceRateLimiter(private val auditLog: AppendOnlyAuditLog) {
    private val timestamps = HashMap<String, MutableList<Long>>()

    @Synchronized
    private fun historyFor(serviceName: String): MutableList<Long> =
        timestamps.getOrPut(serviceName) { mutableListOf() }

    fun tryRollback(serviceName: String, nowMillis: Long): RollbackResult {
        // THE BUG: the read (below) and the eventual write (further down)
        // are not one atomic critical section — another thread can run
        // its own read-decide-write in between.
        val history = historyFor(serviceName)
        val windowStart = nowMillis - WINDOW_MILLIS

        val recent = history.filter { it > windowStart } // O(n) rebuild, every call

        if (recent.size >= MAX_ROLLBACKS_PER_WINDOW) {
            val oldest = recent.min()!!
            val retryAfter = ((oldest + WINDOW_MILLIS - nowMillis).coerceAtLeast(0)) / 1000.0
            return RollbackResult(accepted = false, retryAfterSeconds = retryAfter)
        }

        Thread.sleep(5) // model realistic work between decision and write, widening the race

        synchronized(history) {
            history.clear()
            history.addAll(recent)
            history.add(nowMillis)
        }
        val id = auditLog.append(serviceName, nowMillis)
        return RollbackResult(accepted = true, auditId = id)
    }
}

// ---------------------------------------------------------------------------
// STEP 2 — OPTIMAL
// ---------------------------------------------------------------------------
// Two fixes:
//   1. Efficiency: an ArrayDeque<Long> per service used as a sliding-
//      window log. New timestamps are pushed to the back; expired ones
//      are popped from the front. Since timestamps arrive in
//      non-decreasing order, expired entries are always contiguous at the
//      front -> amortized O(1) eviction (each entry pushed once, popped
//      at most once, ever) instead of an O(n) rebuild per call.
//   2. Correctness: a per-service ReentrantLock makes "evict expired,
//      check size, append" one atomic critical section. A lock PER
//      SERVICE (not one global lock) means a flapping service's
//      rollbacks never add latency to an unrelated service's requests.
// ---------------------------------------------------------------------------
private class ServiceWindow {
    val lock = ReentrantLock()
    val events = ArrayDeque<Long>()
}

class OptimalRateLimiter(private val auditLog: AppendOnlyAuditLog) {
    private val windows = ConcurrentHashMap<String, ServiceWindow>()

    private fun windowFor(serviceName: String): ServiceWindow =
        windows.computeIfAbsent(serviceName) { ServiceWindow() }

    fun tryRollback(serviceName: String, nowMillis: Long): RollbackResult {
        val w = windowFor(serviceName)
        return w.lock.withLock {
            val windowStart = nowMillis - WINDOW_MILLIS
            while (w.events.isNotEmpty() && w.events.peek() <= windowStart) {
                w.events.poll()
            }

            if (w.events.size >= MAX_ROLLBACKS_PER_WINDOW) {
                val oldest = w.events.peek()
                val retryAfter = ((oldest + WINDOW_MILLIS - nowMillis).coerceAtLeast(0)) / 1000.0
                return@withLock RollbackResult(accepted = false, retryAfterSeconds = retryAfter)
            }

            w.events.offer(nowMillis)
            val id = auditLog.append(serviceName, nowMillis)
            RollbackResult(accepted = true, auditId = id)
        }
    }
}
