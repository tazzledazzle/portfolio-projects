/*
 * P5 - Release Lifecycle State Machine
 *
 * Base:  model a release's states and legal transitions; reject illegal ones.
 * Ext 1: an immutable audit trail (who, what, when, why).
 * Ext 2: guards - canary must bake for a minimum duration before ramping.
 * Ext 3: staged rollout percentages, and rollback from any live-ish state.
 * Ext 4: queries - time in state, terminal check, replay of the transition history.
 *
 * The table-driven transition map is the point: adding a state is a data change,
 * not a rewrite of a switch statement.
 */

import java.time.Duration
import java.time.Instant

enum class ReleaseState {
    PENDING, BUILDING, BUILD_FAILED, CANARY, RAMPING, LIVE, ROLLING_BACK, ROLLED_BACK;

    val isTerminal: Boolean get() = this == LIVE || this == BUILD_FAILED || this == ROLLED_BACK
}

enum class Trigger { START_BUILD, BUILD_OK, BUILD_FAIL, PROMOTE_CANARY, RAMP, COMPLETE, ROLLBACK, ROLLBACK_DONE }

data class TransitionRecord(
    val from: ReleaseState,
    val to: ReleaseState,
    val trigger: Trigger,
    val at: Instant,
    val actor: String,
    val note: String? = null
)

class IllegalTransition(val from: ReleaseState, val trigger: Trigger, val why: String) :
    IllegalStateException("cannot $trigger from $from: $why")

class Release(
    val id: String,
    private val clock: ReleaseClock,
    /** canary must bake this long before RAMP is allowed */
    private val minCanaryBake: Duration = Duration.ofMinutes(10),
    private val rampSteps: List<Int> = listOf(5, 25, 50, 100)
) {
    var state: ReleaseState = ReleaseState.PENDING
        private set
    var trafficPercent: Int = 0
        private set

    private val trail = mutableListOf<TransitionRecord>()
    private var enteredStateAt: Instant = clock.now()

    val history: List<TransitionRecord> get() = trail.toList()

    companion object {
        /** The whole policy, in one readable place. */
        private val TABLE: Map<Pair<ReleaseState, Trigger>, ReleaseState> = mapOf(
            (ReleaseState.PENDING to Trigger.START_BUILD) to ReleaseState.BUILDING,
            (ReleaseState.BUILDING to Trigger.BUILD_OK) to ReleaseState.CANARY,
            (ReleaseState.BUILDING to Trigger.BUILD_FAIL) to ReleaseState.BUILD_FAILED,
            (ReleaseState.CANARY to Trigger.RAMP) to ReleaseState.RAMPING,
            (ReleaseState.CANARY to Trigger.ROLLBACK) to ReleaseState.ROLLING_BACK,
            (ReleaseState.RAMPING to Trigger.RAMP) to ReleaseState.RAMPING,
            (ReleaseState.RAMPING to Trigger.COMPLETE) to ReleaseState.LIVE,
            (ReleaseState.RAMPING to Trigger.ROLLBACK) to ReleaseState.ROLLING_BACK,
            (ReleaseState.LIVE to Trigger.ROLLBACK) to ReleaseState.ROLLING_BACK,
            (ReleaseState.ROLLING_BACK to Trigger.ROLLBACK_DONE) to ReleaseState.ROLLED_BACK
        )
    }

    fun allowedTriggers(): Set<Trigger> =
        TABLE.keys.filter { it.first == state }.map { it.second }.toSet()

    fun timeInState(): Duration = Duration.between(enteredStateAt, clock.now())

    fun fire(trigger: Trigger, actor: String, note: String? = null): ReleaseState {
        val target = TABLE[state to trigger]
            ?: throw IllegalTransition(state, trigger, "allowed here: ${allowedTriggers()}")

        guard(trigger)

        val now = clock.now()
        trail += TransitionRecord(state, target, trigger, now, actor, note)
        if (state != target) enteredStateAt = now

        when {
            trigger == Trigger.BUILD_OK -> trafficPercent = rampSteps.first()
            trigger == Trigger.RAMP -> {
                val next = rampSteps.firstOrNull { it > trafficPercent } ?: 100
                trafficPercent = next
                enteredStateAt = now  // each ramp step restarts the bake window
            }
            trigger == Trigger.COMPLETE -> trafficPercent = 100
            trigger == Trigger.ROLLBACK_DONE -> trafficPercent = 0
        }
        state = target
        return state
    }

    /** Ext 2: conditions the transition table alone cannot express. */
    private fun guard(trigger: Trigger) {
        if (trigger == Trigger.RAMP && timeInState() < minCanaryBake) {
            throw IllegalTransition(
                state, trigger,
                "bake time ${timeInState().toMinutes()}m < required ${minCanaryBake.toMinutes()}m"
            )
        }
        if (trigger == Trigger.COMPLETE && trafficPercent < 100) {
            throw IllegalTransition(state, trigger, "still at $trafficPercent% traffic")
        }
    }

    /** Ext 4: total wall-clock time this release has spent in a given state. */
    fun durationIn(target: ReleaseState): Duration {
        var total = Duration.ZERO
        var current = ReleaseState.PENDING
        var since = trail.firstOrNull()?.at ?: clock.now()
        for (record in trail) {
            if (record.from == record.to) continue          // self-loop: same state, keep counting
            if (current == target) total += Duration.between(since, record.at)
            current = record.to
            since = record.at
        }
        if (current == target) total += Duration.between(since, clock.now())
        return total
    }
}

fun interface ReleaseClock { fun now(): Instant }

class TestClock(private var instant: Instant = Instant.parse("2026-08-01T12:00:00Z")) : ReleaseClock {
    override fun now(): Instant = instant
    fun advance(by: Duration) { instant = instant.plus(by) }
}

// ---------------------------------------------------------------- demo / tests

private fun check(label: String, condition: Boolean) {
    if (!condition) throw AssertionError("FAILED: $label")
    println("  ok  $label")
}

private inline fun expectRejection(label: String, body: () -> Unit) {
    try {
        body()
        throw AssertionError("FAILED: $label (expected a rejection)")
    } catch (e: IllegalTransition) {
        println("  ok  $label -> ${e.message}")
    }
}

fun main() {
    println("P5 - Release Lifecycle State Machine")

    val clock = TestClock()
    val release = Release("rel-2026.08.01", clock)

    check("starts pending", release.state == ReleaseState.PENDING)
    check("only START_BUILD is legal first", release.allowedTriggers() == setOf(Trigger.START_BUILD))

    expectRejection("cannot ramp before building") { release.fire(Trigger.RAMP, "terence") }

    release.fire(Trigger.START_BUILD, "release-bot")
    clock.advance(Duration.ofMinutes(4))
    release.fire(Trigger.BUILD_OK, "release-bot")
    check("build success lands in canary", release.state == ReleaseState.CANARY)
    check("canary starts at 5% traffic", release.trafficPercent == 5)

    // Ext 2 - bake guard
    clock.advance(Duration.ofMinutes(3))
    expectRejection("ramp blocked before bake time") { release.fire(Trigger.RAMP, "terence") }

    clock.advance(Duration.ofMinutes(8))
    release.fire(Trigger.RAMP, "terence", note = "error rate flat")
    check("ramped to 25%", release.trafficPercent == 25 && release.state == ReleaseState.RAMPING)

    expectRejection("cannot complete below 100%") { release.fire(Trigger.COMPLETE, "terence") }

    clock.advance(Duration.ofMinutes(11)); release.fire(Trigger.RAMP, "terence")
    clock.advance(Duration.ofMinutes(11)); release.fire(Trigger.RAMP, "terence")
    check("reached 100%", release.trafficPercent == 100)

    release.fire(Trigger.COMPLETE, "terence")
    check("release is live", release.state == ReleaseState.LIVE)
    check("live is terminal-ish but rollback stays open",
        release.state.isTerminal && release.allowedTriggers() == setOf(Trigger.ROLLBACK))

    // Ext 3 - rollback from live
    clock.advance(Duration.ofHours(2))
    release.fire(Trigger.ROLLBACK, "oncall", note = "p99 latency regression")
    release.fire(Trigger.ROLLBACK_DONE, "oncall")
    check("rolled back", release.state == ReleaseState.ROLLED_BACK)
    check("traffic drained", release.trafficPercent == 0)
    check("no triggers remain", release.allowedTriggers().isEmpty())

    // Ext 1 - audit trail
    println("\n  audit trail:")
    release.history.forEach { r ->
        println("    ${r.at} ${r.from} -> ${r.to} via ${r.trigger} by ${r.actor}${r.note?.let { " ($it)" } ?: ""}")
    }
    check("every transition recorded", release.history.size == 8)
    check("trail is chronological",
        release.history.zipWithNext().all { !it.first.at.isAfter(it.second.at) })
    check("rollback reason preserved",
        release.history.any { it.trigger == Trigger.ROLLBACK && it.note == "p99 latency regression" })

    // Ext 4 - queries
    val ramping = release.durationIn(ReleaseState.RAMPING)
    println("  time spent RAMPING = ${ramping.toMinutes()}m")
    check("ramping duration measured", ramping.toMinutes() > 0)

    // A failed build is a dead end.
    val doomed = Release("rel-bad", TestClock())
    doomed.fire(Trigger.START_BUILD, "bot")
    doomed.fire(Trigger.BUILD_FAIL, "bot", note = "compile error in ledger-core")
    check("build failure is terminal", doomed.allowedTriggers().isEmpty() && doomed.state.isTerminal)

    println("P5 passed.\n")
}
