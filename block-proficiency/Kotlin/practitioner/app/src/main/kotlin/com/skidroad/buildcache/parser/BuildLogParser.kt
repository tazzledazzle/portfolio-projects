package com.skidroad.buildcache.parser

import com.skidroad.buildcache.model.*
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow
import kotlinx.coroutines.flow.flowOn
import kotlinx.coroutines.withContext
import java.io.BufferedReader
import java.io.File
import kotlin.time.Duration.Companion.milliseconds

// ── ParseResult — wraps success/failure per line ──────────────────────────

sealed class ParseResult {
    data class Success(val event: BuildEvent) : ParseResult()
    data class Skipped(val line: String, val reason: String) : ParseResult()
    data class Error(val line: String, val cause: Exception) : ParseResult()
}

// ── BuildLogParser ─────────────────────────────────────────────────────────
// Emits a cold Flow<BuildEvent> from a file or stdin.
// Uses flowOn(Dispatchers.IO) so the file read doesn't block the calling coroutine.

class BuildLogParser {

    /**
     * Parse a Gradle scan log file into a cold Flow of BuildEvents.
     * Each line is parsed independently; malformed lines are skipped with a warning.
     *
     * Flow is cold — nothing happens until the caller collects.
     */
    fun parseFile(file: File): Flow<BuildEvent> = flow {
        // withContext(IO) for the file open; the flow itself runs on IO via flowOn below
        val lines = withContext(Dispatchers.IO) {
            file.bufferedReader().use(BufferedReader::readLines)
        }

        for (line in lines) {
            if (line.isBlank() || line.startsWith("#")) continue

            when (val result = parseLine(line)) {
                is ParseResult.Success -> emit(result.event)
                is ParseResult.Skipped -> { /* optionally log */ }
                is ParseResult.Error   -> { /* optionally log */ }
            }
        }
    }.flowOn(Dispatchers.IO)

    /**
     * Parse from stdin — useful for piping: `gradle build 2>&1 | advisor`
     */
    fun parseStdin(): Flow<BuildEvent> = flow {
        val reader = System.`in`.bufferedReader()
        var line: String?
        while (reader.readLine().also { line = it } != null) {
            line?.let { l ->
                if (l.isNotBlank()) {
                    when (val result = parseLine(l)) {
                        is ParseResult.Success -> emit(result.event)
                        else -> { /* skip */ }
                    }
                }
            }
        }
    }.flowOn(Dispatchers.IO)

    // ── Line parsing — Gradle scan TSV format ─────────────────────────────
    // Format: TYPE\tTASK_PATH\tMODULE\tTIMESTAMP\t[fields...]

    private fun parseLine(line: String): ParseResult {
        return try {
            val parts = line.split("\t")
            if (parts.size < 4) return ParseResult.Skipped(line, "too few fields")

            val (type, taskPath, module, rawTimestamp) = parts  // destructuring
            val timestamp = rawTimestamp.toLongOrNull()
                ?: return ParseResult.Skipped(line, "invalid timestamp")

            val event: BuildEvent = when (type.uppercase()) {
                "CACHE_HIT" -> BuildEvent.CacheHit(
                    taskPath = taskPath,
                    module = module,
                    timestamp = timestamp,
                    cacheKey = parts.getOrElse(4) { "unknown" },
                    origin = parts.getOrElse(5) { "LOCAL" }
                        .let { runCatching { CacheOrigin.valueOf(it) }.getOrDefault(CacheOrigin.LOCAL) },
                )
                "CACHE_MISS" -> BuildEvent.CacheMiss(
                    taskPath = taskPath,
                    module = module,
                    timestamp = timestamp,
                    reason = parts.getOrElse(4) { "UNKNOWN" }
                        .let { runCatching { MissReason.valueOf(it) }.getOrDefault(MissReason.UNKNOWN) },
                    executionDuration = parts.getOrElse(5) { "0" }.toLongOrNull()?.milliseconds
                        ?: 0.milliseconds,
                )
                "TASK_SKIPPED" -> BuildEvent.TaskSkipped(
                    taskPath = taskPath,
                    module = module,
                    timestamp = timestamp,
                    skipReason = parts.getOrElse(4) { "UP-TO-DATE" },
                )
                "TASK_FAILED" -> BuildEvent.TaskFailed(
                    taskPath = taskPath,
                    module = module,
                    timestamp = timestamp,
                    errorMessage = parts.getOrElse(4) { "" },
                    exitCode = parts.getOrElse(5) { "-1" }.toIntOrNull() ?: -1,
                )
                "BUILD_STARTED" -> BuildEvent.BuildStarted(
                    taskPath = taskPath,
                    module = module,
                    timestamp = timestamp,
                    gradleVersion = parts.getOrElse(4) { "unknown" },
                    taskCount = parts.getOrElse(5) { "0" }.toIntOrNull() ?: 0,
                )
                "BUILD_FINISHED" -> BuildEvent.BuildFinished(
                    taskPath = taskPath,
                    module = module,
                    timestamp = timestamp,
                    totalDuration = parts.getOrElse(4) { "0" }.toLongOrNull()?.milliseconds
                        ?: 0.milliseconds,
                    outcome = parts.getOrElse(5) { "SUCCESS" }
                        .let { runCatching { BuildOutcome.valueOf(it) }.getOrDefault(BuildOutcome.SUCCESS) },
                )
                else -> return ParseResult.Skipped(line, "unknown type: $type")
            }

            ParseResult.Success(event)
        } catch (e: Exception) {
            ParseResult.Error(line, e)
        }
    }
}