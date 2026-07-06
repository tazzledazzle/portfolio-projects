package com.skidroad.buildcache.cli

import com.github.ajalt.clikt.core.CliktCommand
import com.github.ajalt.clikt.parameters.arguments.argument
import com.github.ajalt.clikt.parameters.arguments.optional
import com.github.ajalt.clikt.parameters.options.default
import com.github.ajalt.clikt.parameters.options.flag
import com.github.ajalt.clikt.parameters.options.option
import com.github.ajalt.clikt.parameters.types.file
import com.skidroad.buildcache.analysis.CacheAnalyzer
import com.skidroad.buildcache.dsl.exceeds
import com.skidroad.buildcache.dsl.report
import com.skidroad.buildcache.model.AnalysisState
import com.skidroad.buildcache.parser.BuildLogParser
import kotlinx.coroutines.*
import kotlinx.coroutines.flow.*
import java.io.File

// ── CLI command ────────────────────────────────────────────────────────────

class AdvisorCommand : CliktCommand(
    name = "advisor",
    help = "Analyze Gradle/Bazel build cache logs and emit prioritized remediation advice.",
) {
    private val logFile: File? by argument(
        name = "LOG_FILE",
        help = "Path to build scan log (TSV format). Reads stdin if omitted.",
    ).file(mustExist = true, canBeDir = false).optional()

    private val onlyHigh: Boolean by option(
        "--only-high",
        help = "Emit only HIGH priority recommendations.",
    ).flag(default = false)

    private val failOnHighMiss: Boolean by option(
        "--fail-on-high-miss",
        help = "Exit with code 1 if overall cache miss rate exceeds threshold.",
    ).flag(default = false)

    private val missThreshold: String by option(
        "--miss-threshold",
        help = "Miss rate threshold for --fail-on-high-miss (default 0.5).",
    ).default("0.5")

    override fun run() {
        // runBlocking bridges the CLI (synchronous) world into coroutines
        runBlocking {
            val exitCode = runAdvisor()
            if (exitCode != 0) throw SystemExit(exitCode)
        }
    }

    private suspend fun runAdvisor(): Int = coroutineScope {
        val parser = BuildLogParser()
        val analyzer = CacheAnalyzer(scope = this)

        // ── 1. Wire up the event flow ──────────────────────────────────────
        val eventFlow: Flow<com.skidroad.buildcache.model.BuildEvent> =
            logFile?.let { parser.parseFile(it) } ?: parser.parseStdin()

        // ── 2. Watch state updates — `launch` for fire-and-forget progress ─
        val progressJob = launch {
            analyzer.state
                .filterNot { it is AnalysisState.Idle }
                .collect { state ->
                    when (state) {
                        is AnalysisState.Parsing ->
                            printProgress("Parsing... ${state.eventsRead} events read")
                        is AnalysisState.Analyzing ->
                            printProgress("Analyzing modules ${state.modulesProcessed}/${state.totalModules} (${state.progressPct}%)")
                        is AnalysisState.Complete ->
                            printProgress("Analysis complete.")
                        is AnalysisState.Failed ->
                            System.err.println("Analysis failed: ${state.cause.message}")
                        is AnalysisState.Idle -> { /* no-op */ }
                    }
                }
        }

        // ── 3. Run analysis — suspends until complete ──────────────────────
        val analysisReport = try {
            analyzer.analyze(eventFlow)
        } catch (e: Exception) {
            System.err.println("Error during analysis: ${e.message}")
            progressJob.cancel()
            return@coroutineScope 2
        }

        progressJob.cancel() // stop watching state after completion

        // ── 4. Render report using the DSL ─────────────────────────────────
        val rendered = report(analysisReport) {
            title("Build Cache Advisor — Analysis Report")

            summary {
                includeWastedTime = true
                includeMissRate   = true
                includeEventCount = true
            }

            section("Module Breakdown (sorted by priority)") {
                metric("Cache Miss Rate") { formatted = true; topN = 20 }
            }

            recommendations {
                onlyHigh = this@AdvisorCommand.onlyHigh
                maxItems = 10
            }
        }

        println(rendered.render())

        // ── 5. Optionally gate CI on miss rate ─────────────────────────────
        val threshold = missThreshold.toDoubleOrNull() ?: 0.5
        return@coroutineScope if (failOnHighMiss && analysisReport.overallCacheMissRate exceeds threshold) {
            System.err.println("FAIL: cache miss rate ${analysisReport.overallCacheMissRate} exceeds threshold $threshold")
            1
        } else {
            0
        }
    }

    private fun printProgress(msg: String) {
        print("\r\u001B[K$msg") // overwrite line
        System.out.flush()
    }
}

// ── Main ───────────────────────────────────────────────────────────────────

fun main(args: Array<String>) = AdvisorCommand().main(args)

// Simple exception to carry exit code out of runBlocking
class SystemExit(val code: Int) : Exception("exit $code")