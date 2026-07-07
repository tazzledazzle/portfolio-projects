package com.skidroad.buildcache.analysis

import com.skidroad.buildcache.model.*
import kotlinx.coroutines.*
import kotlinx.coroutines.flow.*
import kotlin.time.Duration

// ── CacheAnalyzer ──────────────────────────────────────────────────────────
// Consumes a Flow<BuildEvent>, updates a StateFlow<AnalysisState> as work
// progresses, and produces a final AnalysisReport.
//
// Coroutine design:
//   - Collects events on the caller's coroutine scope
//   - Spawns one `async` deferred per module for parallel analysis
//   - Uses `withContext(Default)` for CPU-bound aggregation
//   - Exposes progress via `_state: MutableStateFlow`

class CacheAnalyzer(private val scope: CoroutineScope) {

    private val _state = MutableStateFlow<AnalysisState>(AnalysisState.Idle)

    /** Public read-only view — UI/CLI collects from this. */
    val state: StateFlow<AnalysisState> = _state.asStateFlow()

    /**
     * Collect the full event flow, run per-module analysis in parallel,
     * then synthesize a final report.
     *
     * Emits intermediate [AnalysisState.Parsing] and [AnalysisState.Analyzing]
     * updates so callers can show progress without polling.
     */
    suspend fun analyze(events: Flow<BuildEvent>): AnalysisReport {
        // ── Phase 1: collect all events, emit parse progress ──────────────
        val allEvents = mutableListOf<BuildEvent>()

        events.collect { event ->
            allEvents += event
            _state.value = AnalysisState.Parsing(
                filePath = "stdin",
                eventsRead = allEvents.size,
            )
        }

        // ── Phase 2: group by module, launch parallel analysis ─────────────
        val byModule: Map<String, List<BuildEvent>> = withContext(Dispatchers.Default) {
            allEvents.groupBy { it.module }
        }

        val totalModules = byModule.size
        var modulesProcessed = 0

        _state.value = AnalysisState.Analyzing(
            totalEvents = allEvents.size,
            modulesProcessed = 0,
            totalModules = totalModules,
        )

        // async per module — all run concurrently, awaitAll collects results
        val moduleDeferred: List<Deferred<ModuleAnalysis>> = byModule.map { (module, moduleEvents) ->
            scope.async(Dispatchers.Default) {
                analyzeModule(module, moduleEvents).also {
                    // atomic-ish progress update (slightly racy but fine for display)
                    _state.value = AnalysisState.Analyzing(
                        totalEvents = allEvents.size,
                        modulesProcessed = ++modulesProcessed,
                        totalModules = totalModules,
                    )
                }
            }
        }

        val moduleResults: List<ModuleAnalysis> = moduleDeferred.awaitAll()

        // ── Phase 3: synthesize report ─────────────────────────────────────
        val report = withContext(Dispatchers.Default) {
            synthesizeReport(allEvents, moduleResults)
        }

        _state.value = AnalysisState.Complete(report)
        return report
    }

    // ── Per-module analysis — collections API showcase ────────────────────

    private fun analyzeModule(module: String, events: List<BuildEvent>): ModuleAnalysis {
        val missRate = events.cacheMissRate()
        val wasted = events.totalWastedTime()
        val topReason = events.missReasonBreakdown()
            .entries
            .maxByOrNull { it.value }
            ?.key

        val priority = when {
            missRate > 0.6 || wasted.inWholeMinutes > 5 ->
                RemediationPriority.High("miss rate ${(missRate * 100).toInt()}%")
            missRate > 0.3 || wasted.inWholeMinutes > 1 ->
                RemediationPriority.Medium("miss rate ${(missRate * 100).toInt()}%")
            else ->
                RemediationPriority.Low("miss rate ${(missRate * 100).toInt()}%")
        }

        return ModuleAnalysis(
            module = module,
            events = events,
            cacheMissRate = missRate,
            wastedTime = wasted,
            topMissReason = topReason,
            priority = priority,
        )
    }

    // ── Report synthesis — fold, associate, sortedBy ──────────────────────

    private fun synthesizeReport(
        allEvents: List<BuildEvent>,
        modules: List<ModuleAnalysis>,
    ): AnalysisReport {
        val totalWasted: Duration = modules.fold(Duration.ZERO) { acc, m -> acc + m.wastedTime }

        // associate module name → analysis for O(1) lookup in recommendations
        val moduleMap: Map<String, ModuleAnalysis> = modules.associate { it.module to it }

        val recommendations = buildRecommendations(modules, moduleMap)

        return AnalysisReport(
            totalEvents = allEvents.size,
            overallCacheMissRate = allEvents.cacheMissRate(),
            totalWastedTime = totalWasted,
            moduleBreakdown = modules.sortedBy { it.priority },
            recommendations = recommendations,
        )
    }

    private fun buildRecommendations(
        modules: List<ModuleAnalysis>,
        moduleMap: Map<String, ModuleAnalysis>,
    ): List<Recommendation> {
        val recs = mutableListOf<Recommendation>()

        // Group HIGH modules by their top miss reason → one recommendation per reason
        modules
            .filter { it.priority is RemediationPriority.High }
            .groupBy { it.topMissReason }
            .forEach { (reason, affectedModules) ->
                val (title, detail) = when (reason) {
                    MissReason.INPUTS_CHANGED ->
                        "Unstable task inputs" to
                                "These modules have volatile inputs (timestamps, env vars). " +
                                "Use @InputFiles/@Input annotations and ensure inputs are hermetic."
                    MissReason.NO_CACHE_KEY ->
                        "Tasks not producing cache keys" to
                                "Enable `org.gradle.caching=true` in gradle.properties " +
                                "and annotate tasks with @CacheableTask."
                    MissReason.NOT_CACHEABLE ->
                        "Non-cacheable tasks" to
                                "Mark tasks @CacheableTask or extract non-deterministic " +
                                "logic into lifecycle hooks outside cached tasks."
                    MissReason.CACHE_DISABLED ->
                        "Build cache disabled for these modules" to
                                "Verify `--build-cache` flag is passed in CI and " +
                                "remote cache is reachable from build agents."
                    else ->
                        "Investigate cache misses" to
                                "Review Gradle scan for these modules to diagnose root cause."
                }
                recs += Recommendation(
                    priority = RemediationPriority.High(reason?.name ?: "unknown"),
                    title = title,
                    detail = detail,
                    affectedModules = affectedModules.map { it.module },
                )
            }

        // Remote cache opportunity — modules hitting LOCAL only
        val localOnlyModules = modules.filter { m ->
            m.events
                .filterIsInstance<BuildEvent.CacheHit>()
                .all { it.origin == CacheOrigin.LOCAL }
                .also { _ -> moduleMap[m.module] }
        }
        if (localOnlyModules.isNotEmpty()) {
            recs += Recommendation(
                priority = RemediationPriority.Medium("local-only cache hits"),
                title = "Enable remote cache for cross-agent sharing",
                detail = "These modules hit local cache but miss on fresh CI agents. " +
                        "Configure a remote build cache (Gradle Enterprise, Develocity, or S3 backend).",
                affectedModules = localOnlyModules.map { it.module },
            )
        }

        return recs.sortedBy { it.priority }
    }
}