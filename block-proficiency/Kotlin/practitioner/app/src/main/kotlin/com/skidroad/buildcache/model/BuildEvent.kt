package com.skidroad.buildcache.model

import kotlin.time.Duration

// ── Sealed class hierarchy ─────────────────────────────────────────────────
// Every event that can come from a Gradle/Bazel build log.
// `when` expressions over this hierarchy are exhaustive — no else branch needed.

sealed class BuildEvent {
    abstract val taskPath: String
    abstract val module: String
    abstract val timestamp: Long

    data class CacheHit(
        override val taskPath: String,
        override val module: String,
        override val timestamp: Long,
        val cacheKey: String,
        val origin: CacheOrigin,
    ) : BuildEvent()

    data class CacheMiss(
        override val taskPath: String,
        override val module: String,
        override val timestamp: Long,
        val reason: MissReason,
        val executionDuration: Duration,
    ) : BuildEvent()

    data class TaskSkipped(
        override val taskPath: String,
        override val module: String,
        override val timestamp: Long,
        val skipReason: String,
    ) : BuildEvent()

    data class TaskFailed(
        override val taskPath: String,
        override val module: String,
        override val timestamp: Long,
        val errorMessage: String,
        val exitCode: Int,
    ) : BuildEvent()

    data class BuildStarted(
        override val taskPath: String,
        override val module: String,
        override val timestamp: Long,
        val gradleVersion: String,
        val taskCount: Int,
    ) : BuildEvent()

    data class BuildFinished(
        override val taskPath: String,
        override val module: String,
        override val timestamp: Long,
        val totalDuration: Duration,
        val outcome: BuildOutcome,
    ) : BuildEvent()
}

// ── Supporting enums ───────────────────────────────────────────────────────

enum class CacheOrigin { LOCAL, REMOTE, BOTH }

enum class MissReason {
    NO_CACHE_KEY,
    INPUTS_CHANGED,
    NOT_CACHEABLE,
    CACHE_DISABLED,
    UNKNOWN,
}

enum class BuildOutcome { SUCCESS, FAILURE, ABORTED }

// ── Extension functions on BuildEvent ─────────────────────────────────────

/** Human-readable label for any event — exhaustive when over sealed class. */
fun BuildEvent.label(): String = when (this) {
    is BuildEvent.CacheHit    -> "HIT  [${origin.name}] $taskPath"
    is BuildEvent.CacheMiss   -> "MISS [${reason.name}] $taskPath (${executionDuration})"
    is BuildEvent.TaskSkipped -> "SKIP $taskPath — $skipReason"
    is BuildEvent.TaskFailed  -> "FAIL $taskPath (exit=$exitCode)"
    is BuildEvent.BuildStarted  -> "START gradle=$gradleVersion tasks=$taskCount"
    is BuildEvent.BuildFinished -> "DONE ${outcome.name} in $totalDuration"
}

/** True if this event represents wasted compute. */
fun BuildEvent.isAvoidable(): Boolean = when (this) {
    is BuildEvent.CacheMiss -> reason != MissReason.NOT_CACHEABLE
    else -> false
}

// ── Extension functions on collections of BuildEvent ──────────────────────

fun List<BuildEvent>.cacheMissRate(): Double {
    val actionable = filterIsInstance<BuildEvent.CacheHit>().size +
            filterIsInstance<BuildEvent.CacheMiss>().size
    if (actionable == 0) return 0.0
    return filterIsInstance<BuildEvent.CacheMiss>().size.toDouble() / actionable
}

fun List<BuildEvent>.totalWastedTime(): Duration =
    filterIsInstance<BuildEvent.CacheMiss>()
        .filter { it.isAvoidable() }
        .fold(Duration.ZERO) { acc, miss -> acc + miss.executionDuration }

fun List<BuildEvent>.byModule(): Map<String, List<BuildEvent>> =
    groupBy { it.module }

fun List<BuildEvent>.missReasonBreakdown(): Map<MissReason, Int> =
    filterIsInstance<BuildEvent.CacheMiss>()
        .groupBy { it.reason }
        .mapValues { (_, events) -> events.size }