package com.skidroad.buildcache.dsl

import com.skidroad.buildcache.model.AnalysisReport
import com.skidroad.buildcache.model.Recommendation
import com.skidroad.buildcache.model.RemediationPriority

// ── Report Builder DSL ─────────────────────────────────────────────────────
// Usage:
//   val output = report(analysisReport) {
//       title("Build Cache Analysis")
//       summary { includeWastedTime = true }
//       section("By Module") {
//           metric("Cache Miss Rate") { formatted = true }
//           metric("Wasted Time")     { topN = 5 }
//       }
//       recommendations { onlyHigh = false }
//   }
//
// Lambda receivers: each builder block receives `this` as the relevant builder,
// so callers get autocompletion scoped to that block only.

// ── Top-level entry point ──────────────────────────────────────────────────

fun report(data: AnalysisReport, block: ReportBuilder.() -> Unit): RenderedReport =
    ReportBuilder(data).apply(block).build()

// ── ReportBuilder (receiver for the outer `report { }` block) ─────────────

class ReportBuilder(private val data: AnalysisReport) {
    private var title: String = "Build Cache Advisor Report"
    private val sections = mutableListOf<SectionSpec>()
    private var summarySpec: SummarySpec? = null
    private var recommendationSpec: RecommendationSpec? = null

    fun title(value: String) { title = value }

    /** `summary { }` block — scoped to SummarySpec receiver */
    fun summary(block: SummarySpec.() -> Unit = {}) {
        summarySpec = SummarySpec().apply(block)
    }

    /** `section("name") { }` block — scoped to SectionSpec receiver */
    fun section(name: String, block: SectionSpec.() -> Unit) {
        sections += SectionSpec(name).apply(block)
    }

    /** `recommendations { }` block */
    fun recommendations(block: RecommendationSpec.() -> Unit = {}) {
        recommendationSpec = RecommendationSpec().apply(block)
    }

    internal fun build(): RenderedReport = RenderedReport(
        title = title,
        summary = summarySpec?.render(data),
        sections = sections.map { it.render(data) },
        recommendations = recommendationSpec?.render(data.recommendations),
    )
}

// ── SummarySpec ────────────────────────────────────────────────────────────

class SummarySpec {
    var includeWastedTime: Boolean = true
    var includeMissRate: Boolean = true
    var includeEventCount: Boolean = true

    internal fun render(data: AnalysisReport): String = buildString {
        appendLine("=== Summary ===")
        if (includeEventCount) appendLine("Total events  : ${data.totalEvents}")
        if (includeMissRate)   appendLine("Miss rate     : ${"%.1f".format(data.overallCacheMissRate * 100)}%")
        if (includeWastedTime) appendLine("Wasted compute: ${data.totalWastedTime}")
    }
}

// ── SectionSpec ────────────────────────────────────────────────────────────

class SectionSpec(private val name: String) {
    private val metrics = mutableListOf<MetricSpec>()

    fun metric(name: String, block: MetricSpec.() -> Unit = {}) {
        metrics += MetricSpec(name).apply(block)
    }

    internal fun render(data: AnalysisReport): String = buildString {
        appendLine("=== $name ===")
        data.moduleBreakdown
            .take(metrics.firstOrNull()?.topN ?: Int.MAX_VALUE)
            .forEach { module ->
                appendLine(
                    "  ${module.module.padEnd(40)} " +
                            "miss=${"%.0f".format(module.cacheMissRate * 100)}% " +
                            "wasted=${module.wastedTime} " +
                            "[${module.priority::class.simpleName}]"
                )
            }
    }
}

// ── MetricSpec ─────────────────────────────────────────────────────────────

class MetricSpec(val name: String) {
    var formatted: Boolean = true
    var topN: Int = Int.MAX_VALUE
}

// ── RecommendationSpec ─────────────────────────────────────────────────────

class RecommendationSpec {
    var onlyHigh: Boolean = false
    var maxItems: Int = Int.MAX_VALUE

    internal fun render(recs: List<Recommendation>): String = buildString {
        appendLine("=== Recommendations ===")
        recs
            .let { if (onlyHigh) it.filter { r -> r.priority is RemediationPriority.High } else it }
            .take(maxItems)
            .forEachIndexed { i, rec ->
                appendLine("${i + 1}. [${rec.priority.label}] ${rec.title}")
                appendLine("   ${rec.detail}")
                if (rec.affectedModules.isNotEmpty()) {
                    appendLine("   Affected: ${rec.affectedModules.joinToString(", ")}")
                }
                appendLine()
            }
    }
}

// ── RenderedReport ─────────────────────────────────────────────────────────

data class RenderedReport(
    val title: String,
    val summary: String?,
    val sections: List<String>,
    val recommendations: String?,
) {
    fun render(): String = buildString {
        appendLine("╔══════════════════════════════════════╗")
        appendLine("  $title")
        appendLine("╚══════════════════════════════════════╝")
        appendLine()
        summary?.let { appendLine(it) }
        sections.forEach { appendLine(it) }
        recommendations?.let { appendLine(it) }
    }
}

// ── Infix extension for fluent threshold config ───────────────────────────
// Enables: `missRate exceeds 0.5` in test assertions and config

infix fun Double.exceeds(threshold: Double): Boolean = this > threshold
infix fun Double.within(threshold: Double): Boolean = this <= threshold

// Example usage in policy config (shows infix in a real context):
//
//   if (report.overallCacheMissRate exceeds 0.5) {
//       println("Warning: cache efficiency below threshold")
//   }