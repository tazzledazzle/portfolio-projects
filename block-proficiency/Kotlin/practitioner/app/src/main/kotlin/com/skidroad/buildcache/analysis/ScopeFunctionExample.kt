package com.skidroad.buildcache.analysis

import com.skidroad.buildcache.model.*
import kotlin.time.Duration.Companion.seconds

// ── ScopeFunctionExamples.kt ───────────────────────────────────────────────
// This file is intentionally educational — each scope function is used where
// it's idiomatically *correct*, not just possible.
//
// The rule of thumb:
//   let   → nullable transformation / isolating a value for one expression
//   run   → scoped computation that returns a result (object config + compute)
//   apply → object initialization / builder mutation, returns the receiver
//   also  → side effect without disrupting the chain (logging, validation)
//   with  → operations on an object when you don't need to return it

// ── `let` — nullable transformation ──────────────────────────────────────
// "If this value is non-null, transform it into something else."

fun BuildEvent.CacheMiss.describeWastedTime(): String? =
    executionDuration
        .takeIf { it > 1.seconds }       // null if short
        ?.let { duration ->              // `let` — transform Duration → String
            "Wasted ${duration} on non-cacheable task $taskPath"
        }

// ── `run` — scoped computation that returns a value ───────────────────────
// "Execute a block in the context of an object and return the result."
// Best when you need to configure + immediately compute.

fun buildSampleReport(): AnalysisReport {
    val emptyModules = emptyList<ModuleAnalysis>()

    // `run` on an existing receiver — no temp variable needed
    return emptyModules.run {
        AnalysisReport(
            totalEvents = size,
            overallCacheMissRate = 0.0,
            totalWastedTime = kotlin.time.Duration.ZERO,
            moduleBreakdown = this,
            recommendations = emptyList(),
        )
    }
}

// ── `apply` — object initialization / builder mutation ────────────────────
// "Configure this object and return it." Classic builder pattern.

fun buildSampleRecommendation(): Recommendation =
    Recommendation(
        priority = RemediationPriority.Low("placeholder"),
        title = "",
        detail = "",
        affectedModules = emptyList(),
    ).apply {
        // apply is wrong here in a data class (immutable) — but in a mutable builder it shines.
        // See ReportDsl.kt where `SummarySpec().apply(block)` is the canonical use.
    }

// ── `also` — side effect that doesn't break the chain ─────────────────────
// "Do something with this, then return it unchanged."
// Perfect for logging, validation, or debug instrumentation.

fun List<BuildEvent>.withLogging(label: String): List<BuildEvent> =
    this.also { events ->
        println("[$label] Processing ${events.size} events for ${events.firstOrNull()?.module ?: "unknown"}")
    }

fun CacheAnalyzer.withWarmup(events: List<BuildEvent>): List<BuildEvent> =
    events
        .filter { it !is BuildEvent.BuildStarted }
        .also { println("Filtered to ${it.size} actionable events") }
        .also { require(it.isNotEmpty()) { "No actionable events found" } }

// ── `with` — operate on an object without returning it ────────────────────
// "Call multiple methods on an object; don't need a result."
// Best for printing/rendering blocks where return value is Unit.

fun printModuleSummary(module: ModuleAnalysis) {
    with(module) {
        println("Module  : $module")
        println("Miss %  : ${"%.1f".format(cacheMissRate * 100)}%")
        println("Wasted  : $wastedTime")
        println("Priority: ${priority.label}")
        topMissReason?.let { println("Top miss: ${it.name}") }
    }
}

// ── Chained scope functions — real-world pattern ──────────────────────────
// Shows how they compose naturally without introducing temp vars.

fun List<BuildEvent>.toModuleSummaryLines(): List<String> =
    this
        .filterIsInstance<BuildEvent.CacheMiss>()   // List<CacheMiss>
        .groupBy { it.module }                       // Map<String, List<CacheMiss>>
        .entries
        .sortedByDescending { it.value.size }        // highest miss count first
        .map { (module, misses) ->                   // destructuring in lambda
            misses.run {                             // `run` → compute summary string
                val total = size
                val avoidable = count { it.reason != MissReason.NOT_CACHEABLE }
                "$module: $total misses ($avoidable avoidable)"
            }
        }
        .also { lines ->                             // `also` → log before returning
            println("Generated ${lines.size} module summary lines")
        }