package com.skidroad.buildcache

import com.skidroad.buildcache.analysis.CacheAnalyzer
import com.skidroad.buildcache.dsl.exceeds
import com.skidroad.buildcache.dsl.report
import com.skidroad.buildcache.model.*
import com.skidroad.buildcache.parser.BuildLogParser
import kotlinx.coroutines.flow.flowOf
import kotlinx.coroutines.test.runTest
import java.io.File
import kotlin.test.*
import kotlin.time.Duration.Companion.seconds

class BuildCacheAdvisorTest {

    // ── Sealed class / when exhaustiveness ────────────────────────────────

    @Test
    fun `label() covers all BuildEvent subtypes without else branch`() {
        val events = sampleEvents()
        // If a new subtype is added and label() isn't updated, this won't compile
        val labels = events.map { it.label() }
        assertEquals(events.size, labels.size)
        assertTrue(labels.none { it.isBlank() })
    }

    @Test
    fun `isAvoidable() returns true only for avoidable misses`() {
        val avoidable = BuildEvent.CacheMiss(":compileKotlin", "core", 0L, MissReason.INPUTS_CHANGED, 5.seconds)
        val nonAvoidable = BuildEvent.CacheMiss(":compileKotlin", "core", 0L, MissReason.NOT_CACHEABLE, 5.seconds)
        val hit = BuildEvent.CacheHit(":compileKotlin", "core", 0L, "abc123", CacheOrigin.REMOTE)

        assertTrue(avoidable.isAvoidable())
        assertFalse(nonAvoidable.isAvoidable())
        assertFalse(hit.isAvoidable())
    }

    // ── Collections API ───────────────────────────────────────────────────

    @Test
    fun `cacheMissRate() computes correct ratio`() {
        val events = listOf(
            BuildEvent.CacheHit(":a", "core", 0L, "k1", CacheOrigin.LOCAL),
            BuildEvent.CacheMiss(":b", "core", 0L, MissReason.INPUTS_CHANGED, 1.seconds),
            BuildEvent.CacheMiss(":c", "core", 0L, MissReason.INPUTS_CHANGED, 1.seconds),
        )
        assertEquals(2.0 / 3.0, events.cacheMissRate(), absoluteTolerance = 0.001)
    }

    @Test
    fun `byModule() groups events correctly`() {
        val events = sampleEvents()
        val grouped = events.byModule()
        assertTrue(grouped.containsKey("core"))
        assertTrue(grouped.containsKey("ui"))
    }

    @Test
    fun `missReasonBreakdown() aggregates counts`() {
        val events = listOf(
            BuildEvent.CacheMiss(":a", "m", 0L, MissReason.INPUTS_CHANGED, 1.seconds),
            BuildEvent.CacheMiss(":b", "m", 0L, MissReason.INPUTS_CHANGED, 1.seconds),
            BuildEvent.CacheMiss(":c", "m", 0L, MissReason.NO_CACHE_KEY, 1.seconds),
        )
        val breakdown = events.missReasonBreakdown()
        assertEquals(2, breakdown[MissReason.INPUTS_CHANGED])
        assertEquals(1, breakdown[MissReason.NO_CACHE_KEY])
    }

    // ── Coroutines / Flow ─────────────────────────────────────────────────

    @Test
    fun `analyzer emits Complete state and produces report`() = runTest {
        val analyzer = CacheAnalyzer(scope = this)
        val events = flowOf(*sampleEvents().toTypedArray())

        val report = analyzer.analyze(events)

        assertIs<AnalysisState.Complete>(analyzer.state.value)
        assertTrue(report.totalEvents > 0)
        assertTrue(report.moduleBreakdown.isNotEmpty())
    }

    @Test
    fun `parser emits correct number of events from test file`() = runTest {
        val testLog = writeTempLog(
            "CACHE_HIT\t:compileKotlin\tcore\t1000\tabc\tLOCAL",
            "CACHE_MISS\t:test\tcore\t2000\tINPUTS_CHANGED\t5000",
            "TASK_SKIPPED\t:check\tui\t3000\tUP-TO-DATE",
        )

        val parser = BuildLogParser()
        val collected = mutableListOf<BuildEvent>()
        parser.parseFile(testLog).collect { collected += it }

        assertEquals(3, collected.size)
        assertIs<BuildEvent.CacheHit>(collected[0])
        assertIs<BuildEvent.CacheMiss>(collected[1])
        assertIs<BuildEvent.TaskSkipped>(collected[2])

        testLog.delete()
    }

    // ── DSL ───────────────────────────────────────────────────────────────

    @Test
    fun `report DSL renders without throwing`() = runTest {
        val analyzer = CacheAnalyzer(scope = this)
        val analysisReport = analyzer.analyze(flowOf(*sampleEvents().toTypedArray()))

        val rendered = report(analysisReport) {
            title("Test Report")
            summary { includeWastedTime = true }
            section("Modules") { metric("miss rate") }
            recommendations { onlyHigh = true }
        }

        val output = rendered.render()
        assertTrue(output.contains("Test Report"))
    }

    // ── Infix functions ───────────────────────────────────────────────────

    @Test
    fun `exceeds infix returns correct boolean`() {
        assertTrue(0.6 exceeds 0.5)
        assertFalse(0.4 exceeds 0.5)
    }

    // ── Helpers ───────────────────────────────────────────────────────────

    private fun sampleEvents(): List<BuildEvent> = listOf(
        BuildEvent.CacheHit(":compileKotlin", "core", 1000L, "k1", CacheOrigin.LOCAL),
        BuildEvent.CacheMiss(":test", "core", 2000L, MissReason.INPUTS_CHANGED, 10.seconds),
        BuildEvent.CacheMiss(":lint", "core", 3000L, MissReason.INPUTS_CHANGED, 5.seconds),
        BuildEvent.CacheHit(":compileKotlin", "ui", 4000L, "k2", CacheOrigin.REMOTE),
        BuildEvent.TaskSkipped(":check", "ui", 5000L, "UP-TO-DATE"),
        BuildEvent.TaskFailed(":integrationTest", "core", 6000L, "OutOfMemoryError", 1),
        BuildEvent.BuildStarted(":build", "root", 500L, "8.6", 42),
        BuildEvent.BuildFinished(":build", "root", 9000L, 8.5.seconds, BuildOutcome.FAILURE),
    )

    private fun writeTempLog(vararg lines: String): File =
        File.createTempFile("build-log-", ".tsv").apply {
            writeText(lines.joinToString("\n"))
            deleteOnExit()
        }
}