/*
 * P4 - Idempotent Deploy-Event Processor
 *
 * The scenario: your CI provider delivers webhooks at-least-once. Retries, duplicates,
 * and reordering are normal. Process each logical event exactly once.
 *
 * Base:  dedupe by idempotency key, return the cached result on a replay.
 * Ext 1: failures are NOT cached - a retry after an error must actually re-run.
 * Ext 2: TTL eviction on the dedupe store (an injected clock keeps this testable).
 * Ext 3: concurrent duplicates - the second caller must not run the body twice.
 * Ext 4: out-of-order versioning - a stale update for a key is dropped, not applied.
 */

import java.time.Duration
import java.time.Instant
import java.util.concurrent.ConcurrentHashMap
import java.util.concurrent.CountDownLatch
import java.util.concurrent.Executors
import java.util.concurrent.TimeUnit
import java.util.concurrent.atomic.AtomicInteger

fun interface Clock {
    fun now(): Instant
}

class ManualClock(private var instant: Instant = Instant.parse("2026-08-01T00:00:00Z")) : Clock {
    override fun now(): Instant = instant
    fun advance(by: Duration) { instant = instant.plus(by) }
}

sealed interface Outcome<out R> {
    /** Ran the body for the first time. */
    data class Processed<R>(val result: R) : Outcome<R>
    /** Seen before; replayed the stored result without re-running. */
    data class Duplicate<R>(val result: R) : Outcome<R>
    /** A newer version for this key already landed. */
    data class Stale(val seenVersion: Long, val submittedVersion: Long) : Outcome<Nothing>
    /** Body threw; nothing cached, so the caller may retry. */
    data class Failed(val error: Throwable) : Outcome<Nothing>
}

private class Record<R>(val version: Long, val result: R, val storedAt: Instant)

class IdempotentProcessor<R>(
    private val clock: Clock = Clock { Instant.now() },
    private val ttl: Duration = Duration.ofHours(24)
) {
    private val committed = ConcurrentHashMap<String, Record<R>>()
    /** per-key mutex so two threads with the same key serialize instead of racing */
    private val locks = ConcurrentHashMap<String, Any>()

    val size: Int get() = committed.size

    fun process(key: String, version: Long = 0, body: () -> R): Outcome<R> {
        // Fast path: no lock needed for a clean replay.
        liveRecord(key)?.let { existing ->
            if (version < existing.version) return Outcome.Stale(existing.version, version)
            if (version == existing.version) return Outcome.Duplicate(existing.result)
        }

        val lock = locks.computeIfAbsent(key) { Any() }
        synchronized(lock) {
            // Re-check under the lock: another thread may have committed while we waited.
            liveRecord(key)?.let { existing ->
                if (version < existing.version) return Outcome.Stale(existing.version, version)
                if (version == existing.version) return Outcome.Duplicate(existing.result)
            }
            return try {
                val result = body()
                committed[key] = Record(version, result, clock.now())
                Outcome.Processed(result)
            } catch (t: Throwable) {
                // Deliberately do not cache: a transient failure must stay retryable.
                Outcome.Failed(t)
            } finally {
                locks.remove(key, lock)
            }
        }
    }

    /** Returns the record only if it has not aged out. Lazily evicts expired entries. */
    private fun liveRecord(key: String): Record<R>? {
        val record = committed[key] ?: return null
        return if (Duration.between(record.storedAt, clock.now()) >= ttl) {
            committed.remove(key, record); null
        } else record
    }

    fun evictExpired(): Int {
        val now = clock.now()
        val doomed = committed.entries.filter { Duration.between(it.value.storedAt, now) >= ttl }
        doomed.forEach { committed.remove(it.key, it.value) }
        return doomed.size
    }
}

// ---------------------------------------------------------------- demo / tests

private fun check(label: String, condition: Boolean) {
    if (!condition) throw AssertionError("FAILED: $label")
    println("  ok  $label")
}

fun main() {
    println("P4 - Idempotent Deploy-Event Processor")

    val clock = ManualClock()
    val processor = IdempotentProcessor<String>(clock, ttl = Duration.ofMinutes(30))
    val runs = AtomicInteger(0)

    fun deploy(id: String) = { runs.incrementAndGet(); "deployed:$id" }

    val first = processor.process("evt-1", body = deploy("ledger"))
    val replay = processor.process("evt-1", body = deploy("ledger"))
    println("first  = $first")
    println("replay = $replay")

    check("first call processes", first is Outcome.Processed && first.result == "deployed:ledger")
    check("replay is a duplicate", replay is Outcome.Duplicate)
    check("body ran exactly once", runs.get() == 1)

    // Ext 1 - failures stay retryable
    val flaky = IdempotentProcessor<String>(clock)
    val attempts = AtomicInteger(0)
    val body = {
        if (attempts.incrementAndGet() < 3) throw IllegalStateException("registry timeout")
        "ok after ${attempts.get()}"
    }
    val r1 = flaky.process("evt-2", body = body)
    val r2 = flaky.process("evt-2", body = body)
    val r3 = flaky.process("evt-2", body = body)
    val r4 = flaky.process("evt-2", body = body)
    println("retries: ${r1::class.simpleName}, ${r2::class.simpleName}, ${r3::class.simpleName}, ${r4::class.simpleName}")
    check("first two attempts fail", r1 is Outcome.Failed && r2 is Outcome.Failed)
    check("third attempt succeeds", r3 is Outcome.Processed)
    check("fourth is a duplicate", r4 is Outcome.Duplicate)
    check("failures were not cached", attempts.get() == 3)

    // Ext 2 - TTL
    check("key is still cached before TTL", processor.process("evt-1") { "x" } is Outcome.Duplicate)
    clock.advance(Duration.ofMinutes(31))
    val afterTtl = processor.process("evt-1", body = deploy("ledger"))
    println("after TTL = $afterTtl")
    check("expired key reprocesses", afterTtl is Outcome.Processed)
    check("body ran a second time", runs.get() == 2)

    // Ext 4 - out-of-order versions
    val versioned = IdempotentProcessor<String>(ManualClock())
    versioned.process("svc-a", version = 5) { "v5" }
    val stale = versioned.process("svc-a", version = 3) { "v3" }
    val newer = versioned.process("svc-a", version = 6) { "v6" }
    println("stale = $stale")
    check("older version is rejected", stale is Outcome.Stale && stale.seenVersion == 5L)
    check("newer version is applied", newer is Outcome.Processed && newer.result == "v6")

    // Ext 3 - concurrent duplicates
    val racy = IdempotentProcessor<String>(ManualClock())
    val bodyRuns = AtomicInteger(0)
    val threads = 16
    val gun = CountDownLatch(1)
    val pool = Executors.newFixedThreadPool(threads)
    val outcomes = java.util.concurrent.ConcurrentLinkedQueue<Outcome<String>>()
    repeat(threads) {
        pool.submit {
            gun.await()
            outcomes += racy.process("hot-key") {
                bodyRuns.incrementAndGet()
                Thread.sleep(5)
                "single-result"
            }
        }
    }
    gun.countDown()
    pool.shutdown()
    pool.awaitTermination(10, TimeUnit.SECONDS)

    val processed = outcomes.count { it is Outcome.Processed }
    val duplicates = outcomes.count { it is Outcome.Duplicate }
    println("under $threads concurrent callers: processed=$processed duplicates=$duplicates bodyRuns=${bodyRuns.get()}")
    check("body ran once under contention", bodyRuns.get() == 1)
    check("exactly one caller processed", processed == 1)
    check("everyone else saw a duplicate", duplicates == threads - 1)

    println("P4 passed.\n")
}
