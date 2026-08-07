package org.example


import java.time.Duration
import java.time.Instant
import java.util.ArrayDeque
import java.util.PriorityQueue
import java.util.concurrent.*
import java.util.concurrent.atomic.AtomicInteger

// ============ 5. FAILURE HANDLING ============

// Result / runCatching: for recoverable failures you want to inspect
fun fetchLimit(service: String): Result<Int> = runCatching {
    if (service == "down") error("upstream 503") else 5000
}

// Sealed error type: better than Result when callers must branch on the *kind*
sealed interface ChargeError {
    data class OverLimit(val requested: Int, val available: Int) : ChargeError
    data object CardFrozen : ChargeError
    data class Upstream(val cause: String) : ChargeError
}

sealed interface Charge {
    data class Approved(val id: String, val remaining: Int) : Charge
    data class Declined(val error: ChargeError) : Charge
}

class CreditLimitChecker(private val limits: Map<String, Int>, private val frozen: Set<String> = emptySet()) {
    private val spent = mutableMapOf<String, Int>()

    fun charge(card: String, amountCents: Int): Charge {
        require(amountCents > 0) { "amount must be positive, got $amountCents" }
        if (card in frozen) return Charge.Declined(ChargeError.CardFrozen)
        val limit = limits[card] ?: return Charge.Declined(ChargeError.Upstream("unknown card $card"))
        val used = spent.getOrDefault(card, 0)
        val available = limit - used
        if (amountCents > available) return Charge.Declined(ChargeError.OverLimit(amountCents, available))
        spent[card] = used + amountCents
        return Charge.Approved("txn-${card}-${used + amountCents}", available - amountCents)
    }
}

// Retry with exponential backoff + jitter, only on retryable failures
fun <T> retry(
    attempts: Int = 3,
    initialDelay: Duration = Duration.ofMillis(10),
    retryOn: (Throwable) -> Boolean = { true },
    sleep: (Duration) -> Unit = { Thread.sleep(it.toMillis()) },
    body: () -> T
): Result<T> {
    var delay = initialDelay
    repeat(attempts - 1) {
        val r = runCatching(body)
        if (r.isSuccess) return r
        val e = r.exceptionOrNull()!!
        if (!retryOn(e)) return r
        sleep(delay)
        delay = delay.multipliedBy(2)
    }
    return runCatching(body)
}

// ============ 6. IDEMPOTENCY ============

fun interface Clock { fun now(): Instant }
class FakeClock(private var t: Instant = Instant.parse("2026-08-01T00:00:00Z")) : Clock {
    override fun now() = t
    fun advance(d: Duration) { t = t.plus(d) }
}

sealed interface Applied<out R> {
    data class First<R>(val result: R) : Applied<R>
    data class Replay<R>(val result: R) : Applied<R>
    data class Failed(val cause: Throwable) : Applied<Nothing>
}

class IdempotencyStore<R>(private val clock: Clock, private val ttl: Duration = Duration.ofHours(24)) {
    private class Entry<R>(val result: R, val at: Instant)

    private val committed = ConcurrentHashMap<String, Entry<R>>()
    private val locks = ConcurrentHashMap<String, Any>()

    fun once(key: String, body: () -> R): Applied<R> {
        live(key)?.let { return Applied.Replay(it.result) }          // fast path, no lock
        val lock = locks.computeIfAbsent(key) { Any() }
        synchronized(lock) {
            live(key)?.let { return Applied.Replay(it.result) }      // re-check under lock
            return try {
                val r = body()
                committed[key] = Entry(r, clock.now())               // commit ONLY on success
                Applied.First(r)
            } catch (t: Throwable) {
                Applied.Failed(t)                                    // never cache a failure
            } finally {
                locks.remove(key, lock)
            }
        }
    }

    private fun live(key: String): Entry<R>? {
        val e = committed[key] ?: return null
        return if (Duration.between(e.at, clock.now()) >= ttl) { committed.remove(key, e); null } else e
    }
}

// ============ 7. SCHEDULING STRUCTURES ============

data class Task(val name: String, val priority: Int, val runAt: Instant, val deps: List<String> = emptyList())

fun priorityQueueDemo(tasks: List<Task>): List<String> {
    // lowest priority number first, then earliest runAt, then name for determinism
    val cmp = compareBy<Task>({ it.priority }, { it.runAt }, { it.name })
    val pq = PriorityQueue(cmp).apply { addAll(tasks) }
    return generateSequence { pq.poll() }.map { it.name }.toList()
}

fun topoSort(tasks: List<Task>): List<List<String>>? {
    val byName = tasks.associateBy { it.name }
    val dependents = mutableMapOf<String, MutableList<String>>()
    tasks.forEach { t -> t.deps.forEach { dependents.getOrPut(it) { mutableListOf() }.add(t.name) } }
    val indeg = byName.mapValues { it.value.deps.size }.toMutableMap()
    var frontier = indeg.filterValues { it == 0 }.keys.sorted()
    val waves = mutableListOf<List<String>>()
    var seen = 0
    while (frontier.isNotEmpty()) {
        waves += frontier; seen += frontier.size
        val next = mutableListOf<String>()
        frontier.forEach { d ->
            dependents[d].orEmpty().forEach { c ->
                indeg[c] = indeg.getValue(c) - 1
                if (indeg.getValue(c) == 0) next += c
            }
        }
        frontier = next.sorted()
    }
    return if (seen == byName.size) waves else null   // null == cycle
}

// BFS/queue: ArrayDeque is the Kotlin idiom for both stack and queue
fun bfsLevels(adj: Map<String, List<String>>, start: String): Map<String, Int> {
    val dist = mutableMapOf(start to 0)
    val q = ArrayDeque(listOf(start))
    while (q.isNotEmpty()) {
        val n = q.removeFirst()
        adj[n].orEmpty().forEach { m -> if (m !in dist) { dist[m] = dist.getValue(n) + 1; q.addLast(m) } }
    }
    return dist
}

// interval merge: the other scheduling classic (maintenance windows, freezes)
fun mergeWindows(windows: List<Pair<Instant, Instant>>): List<Pair<Instant, Instant>> {
    if (windows.isEmpty()) return emptyList()
    val out = mutableListOf<Pair<Instant, Instant>>()
    windows.sortedBy { it.first }.forEach { (s, e) ->
        val last = out.lastOrNull()
        if (last != null && !s.isAfter(last.second)) out[out.lastIndex] = last.first to maxOf(last.second, e)
        else out += s to e
    }
    return out
}

// ============ 8. STATE MACHINE ============

enum class S { NEW, RUNNING, PAUSED, DONE, FAILED }
enum class E { START, PAUSE, RESUME, FINISH, FAIL }

class Machine(private val clock: Clock) {
    var state: S = S.NEW; private set
    val log = mutableListOf<Triple<S, E, S>>()

    private val table: Map<Pair<S, E>, S> = mapOf(
        (S.NEW to E.START) to S.RUNNING,
        (S.RUNNING to E.PAUSE) to S.PAUSED,
        (S.PAUSED to E.RESUME) to S.RUNNING,
        (S.RUNNING to E.FINISH) to S.DONE,
        (S.RUNNING to E.FAIL) to S.FAILED,
        (S.PAUSED to E.FAIL) to S.FAILED
    )

    private var enteredAt = clock.now()
    fun allowed(): Set<E> = table.keys.filter { it.first == state }.map { it.second }.toSet()
    fun timeInState(): Duration = Duration.between(enteredAt, clock.now())

    fun fire(e: E): S {
        val next = table[state to e] ?: throw IllegalStateException("cannot $e from $state (allowed: ${allowed()})")
        log += Triple(state, e, next)
        state = next; enteredAt = clock.now()
        return next
    }
}

// ============ TESTS ============

fun main() {
    // credit limit checker
    val checker = CreditLimitChecker(mapOf("card-1" to 10_000), frozen = setOf("card-frozen"))
    check(checker.charge("card-1", 3_000) is Charge.Approved)
    val over = checker.charge("card-1", 9_000)
    check(over is Charge.Declined && over.error == ChargeError.OverLimit(9_000, 7_000)) { over.toString() }
    check((checker.charge("card-frozen", 1) as Charge.Declined).error == ChargeError.CardFrozen)
    check(runCatching { checker.charge("card-1", -5) }.isFailure)

    check(fetchLimit("ledger").getOrNull() == 5000)
    check(fetchLimit("down").isFailure)
    check(fetchLimit("down").getOrDefault(0) == 0)

    // retry
    val calls = AtomicInteger(0)
    val r = retry(attempts = 4, sleep = {}) {
        if (calls.incrementAndGet() < 3) error("flaky") else "ok"
    }
    check(r.getOrNull() == "ok" && calls.get() == 3)

    val nonRetryable = AtomicInteger(0)
    val r2 = retry(attempts = 5, sleep = {}, retryOn = { it !is IllegalArgumentException }) {
        nonRetryable.incrementAndGet(); throw IllegalArgumentException("bad input")
    }
    check(r2.isFailure && nonRetryable.get() == 1) { "tried ${nonRetryable.get()}" }

    // idempotency
    val clock = FakeClock()
    val store = IdempotencyStore<String>(clock, ttl = Duration.ofMinutes(30))
    val runs = AtomicInteger(0)
    val a = store.once("k1") { runs.incrementAndGet(); "v" }
    val b = store.once("k1") { runs.incrementAndGet(); "v" }
    check(a is Applied.First && b is Applied.Replay && runs.get() == 1)

    val fails = AtomicInteger(0)
    check(store.once("k2") { fails.incrementAndGet(); error("boom") } is Applied.Failed)
    check(store.once("k2") { fails.incrementAndGet(); "recovered" } is Applied.First)
    check(fails.get() == 2) { "failure was cached" }

    clock.advance(Duration.ofMinutes(31))
    check(store.once("k1") { runs.incrementAndGet(); "v2" } is Applied.First && runs.get() == 2)

    // concurrent duplicates
    val hot = IdempotencyStore<String>(FakeClock())
    val bodyRuns = AtomicInteger(0)
    val gate = CountDownLatch(1)
    val pool = Executors.newFixedThreadPool(16)
    val results = ConcurrentLinkedQueue<Applied<String>>()
    repeat(16) { pool.submit { gate.await(); results += hot.once("hot") { bodyRuns.incrementAndGet(); Thread.sleep(5); "one" } } }
    gate.countDown(); pool.shutdown(); pool.awaitTermination(10, TimeUnit.SECONDS)
    check(bodyRuns.get() == 1) { "ran ${bodyRuns.get()} times" }
    check(results.count { it is Applied.First } == 1)

    // scheduling
    val t0 = Instant.parse("2026-08-01T00:00:00Z")
    val tasks = listOf(
        Task("build", 1, t0),
        Task("test", 1, t0.plusSeconds(60), deps = listOf("build")),
        Task("deploy", 0, t0.plusSeconds(120), deps = listOf("test")),
        Task("lint", 2, t0)
    )
    check(priorityQueueDemo(tasks).first() == "deploy")
    check(topoSort(tasks) == listOf(listOf("build", "lint"), listOf("test"), listOf("deploy")))
    check(topoSort(listOf(Task("a", 0, t0, listOf("b")), Task("b", 0, t0, listOf("a")))) == null)

    check(bfsLevels(mapOf("a" to listOf("b", "c"), "b" to listOf("d")), "a") ==
        mapOf("a" to 0, "b" to 1, "c" to 1, "d" to 2))

    val merged = mergeWindows(listOf(
        t0 to t0.plusSeconds(100),
        t0.plusSeconds(50) to t0.plusSeconds(150),
        t0.plusSeconds(300) to t0.plusSeconds(400)
    ))
    check(merged.size == 2 && merged[0].second == t0.plusSeconds(150))

    // state machine
    val m = Machine(FakeClock())
    check(m.allowed() == setOf(E.START))
    check(runCatching { m.fire(E.FINISH) }.isFailure)
    m.fire(E.START); m.fire(E.PAUSE); m.fire(E.RESUME); m.fire(E.FINISH)
    check(m.state == S.DONE && m.allowed().isEmpty() && m.log.size == 4)

    println("File B: all checks passed.")
}
