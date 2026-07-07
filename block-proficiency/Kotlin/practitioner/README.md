# Build Cache Advisor

A CLI tool that ingests Gradle/Bazel build scan logs, streams events reactively,
analyzes cache efficiency per module, and emits prioritized remediation advice.

Born from production work at Tableau (Bazel migration) and Invisible Technologies
(280+ microservice CI pipeline). This standalone project distills that experience
into a clean Kotlin showcase.

---

## Run

```bash
# From a log file
./gradlew run --args="path/to/build-scan.tsv"

# From stdin (pipe from Gradle)
gradle build 2>&1 | ./gradlew run

# Fail CI if miss rate > 40%
./gradlew run --args="scan.tsv --fail-on-high-miss --miss-threshold 0.4"

# Only show HIGH priority recommendations
./gradlew run --args="scan.tsv --only-high"

# Run tests
./gradlew test
```

---

## Log Format

Tab-separated values, one event per line:

```
TYPE\tTASK_PATH\tMODULE\tTIMESTAMP_MS\t[field1]\t[field2]
```

| Type | Fields |
|---|---|
| `CACHE_HIT` | cacheKey, origin (LOCAL\|REMOTE\|BOTH) |
| `CACHE_MISS` | reason (INPUTS_CHANGED\|NO_CACHE_KEY\|NOT_CACHEABLE\|CACHE_DISABLED), durationMs |
| `TASK_SKIPPED` | skipReason |
| `TASK_FAILED` | errorMessage, exitCode |
| `BUILD_STARTED` | gradleVersion, taskCount |
| `BUILD_FINISHED` | durationMs, outcome (SUCCESS\|FAILURE\|ABORTED) |

Lines starting with `#` are comments. Blank lines are ignored.

---

## Kotlin Capabilities — Where to Find Each

| Capability | Location | Notes |
|---|---|---|
| **Sealed classes + `when`** | `model/BuildEvent.kt` | Full hierarchy; `label()` is exhaustive — no `else` |
| **Destructuring** | `parser/BuildLogParser.kt` | `val (type, taskPath, module, rawTimestamp) = parts` |
| **`Flow` (cold)** | `parser/BuildLogParser.kt` | `parseFile()` returns `Flow<BuildEvent>` |
| **`StateFlow`** | `analysis/CacheAnalyzer.kt` | `_state: MutableStateFlow<AnalysisState>` |
| **`launch` / `async`** | `analysis/CacheAnalyzer.kt` | `async` per module, `awaitAll()`, progress via `launch` |
| **`suspend` + `withContext`** | `analysis/CacheAnalyzer.kt` | `withContext(Dispatchers.Default)` for CPU-bound work |
| **`let run apply also with`** | `analysis/ScopeFunctionShowcase.kt` | Each function used idiomatically with explanation |
| **Collections API** | `model/BuildEvent.kt`, `analysis/CacheAnalyzer.kt` | `groupBy`, `associate`, `fold`, `filter`, `map`, `partition` |
| **DSL (lambda receivers)** | `dsl/ReportDsl.kt` | `report { section { metric { } } }` |
| **Infix functions** | `dsl/ReportDsl.kt` | `missRate exceeds 0.5` |

---

## Architecture

```
stdin / file
     │
     ▼
BuildLogParser          (Flow<BuildEvent>)
     │
     ▼
CacheAnalyzer           (StateFlow<AnalysisState> for live progress)
  ├── async per module  (parallel ModuleAnalysis)
  └── synthesizeReport  (fold, associate, groupBy)
     │
     ▼
ReportDSL               (lambda receiver DSL → RenderedReport)
     │
     ▼
CLI stdout / exit code
```

---

## Design Decisions

**Why `Flow` instead of collecting into a list first?**
Build scan logs can be large (100K+ events for a 280-service monorepo). Cold `Flow`
lets the parser emit events lazily without holding the full log in memory.

**Why `StateFlow` for progress, not callbacks?**
`StateFlow` is observable from multiple collectors (CLI progress renderer +
potential future web UI) without any coupling between them. It also replays
the latest state to late subscribers automatically.

**Why `async/awaitAll` for module analysis?**
Module analyses are independent and CPU-bound. `async` on `Dispatchers.Default`
saturates available cores without blocking the main coroutine. With 100+ modules
this is measurably faster than sequential analysis.