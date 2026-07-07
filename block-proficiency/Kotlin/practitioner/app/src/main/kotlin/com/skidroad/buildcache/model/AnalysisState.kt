package com.skidroad.buildcache.model

// ── Analysis progress — StateFlow payload ─────────────────────────────────
// Emitted via StateFlow<AnalysisState> so the CLI can render live progress.

sealed class AnalysisState {
    data object Idle : AnalysisState()

    data class Parsing(
        val filePath: String,
        val eventsRead: Int,
    ) : AnalysisState()

    data class Analyzing(
        val totalEvents: Int,
        val modulesProcessed: Int,
        val totalModules: Int,
    ) : AnalysisState() {
        val progressPct: Int get() = if (totalModules == 0) 0
        else (modulesProcessed * 100) / totalModules
    }

    data class Complete(val report: AnalysisReport) : AnalysisState()

    data class Failed(val cause: Throwable) : AnalysisState()
}

// ── Remediation priority — drives report ordering ─────────────────────────

sealed class RemediationPriority : Comparable<RemediationPriority> {
    data class High(val reason: String)   : RemediationPriority()
    data class Medium(val reason: String) : RemediationPriority()
    data class Low(val reason: String)    : RemediationPriority()

    private val ordinal: Int get() = when (this) {
        is High   -> 0
        is Medium -> 1
        is Low    -> 2
    }

    override fun compareTo(other: RemediationPriority): Int =
        this.ordinal.compareTo(other.ordinal)

    val label: String get() = when (this) {
        is High   -> "HIGH   — $reason"
        is Medium -> "MEDIUM — $reason"
        is Low    -> "LOW    — $reason"
    }
}

// ── Per-module analysis result ─────────────────────────────────────────────

data class ModuleAnalysis(
    val module: String,
    val events: List<BuildEvent>,
    val cacheMissRate: Double,
    val wastedTime: kotlin.time.Duration,
    val topMissReason: MissReason?,
    val priority: RemediationPriority,
)

// ── Top-level report ───────────────────────────────────────────────────────

data class AnalysisReport(
    val totalEvents: Int,
    val overallCacheMissRate: Double,
    val totalWastedTime: kotlin.time.Duration,
    val moduleBreakdown: List<ModuleAnalysis>,
    val recommendations: List<Recommendation>,
)

data class Recommendation(
    val priority: RemediationPriority,
    val title: String,
    val detail: String,
    val affectedModules: List<String>,
)