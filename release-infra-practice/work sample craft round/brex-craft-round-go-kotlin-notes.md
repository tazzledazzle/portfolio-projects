# Brex Craft-Round Problems — Go & Kotlin Ports

Companion to `brex-craft-round-walkthroughs.md`. That doc has the full
brute-force → optimal reasoning, complexity analysis, and test matrices for
all 5 problems — none of that changes by language, so it isn't repeated
here. This doc covers what actually **is** language-specific: which JVM/Go
primitive replaces which Python one, and a few real bugs the compiler/test
run caught along the way that are worth knowing about.

**All three language ports were compiled and test-run in this environment,
not just written:**

|                                 | Python                                   | Go                       | Kotlin        |
|---------------------------------|------------------------------------------|--------------------------|---------------|
| Tests                           | 72                                       | 52                       | 53            |
| Result                          | all pass                                 | all pass (`-race` clean) | all pass      |
| Diamond-graph blowup (depth 16) | 98,303 brute-force calls vs. 32 memoized | 98,303 vs. 32            | 98,303 vs. 32 |

That last row is a nice thing to say out loud in an interview if it comes
up: the algorithmic behavior is identical across three completely different
runtimes, because the bug (and the fix) is structural, not language-specific.

```
brex_solutions_go/         brex_solutions_kotlin/
├── go.mod                 ├── p1/{p1.kt, P1Test.kt, RunTests.kt}
├── p1/{p1.go, p1_test.go} ├── p2/{p2.kt, P2Test.kt, RunTests.kt}
├── p2/...                 ├── p3/...
├── p3/...                 ├── p4/...
├── p4/...                 └── p5/...
└── p5/...
```

## Running them

**Go** — standard tooling, no setup beyond a Go install:
```bash
cd brex_solutions_go
go test ./...              # all packages
go test ./p1/... -race -v  # race detector, verbose, one package
```

**Kotlin** — these were compiled directly with `kotlinc` + JUnit4 (no
Gradle/Maven project scaffolding, to keep the deliverable to plain files
you can drop into any project). Each problem is self-contained; compile and
run one at a time:
```bash
cd brex_solutions_kotlin/p1
kotlinc p1.kt P1Test.kt RunTests.kt -cp /path/to/junit4.jar -include-runtime -d p1.jar
java -cp "p1.jar:/path/to/junit4.jar" RunTestsKt
```
`RunTests.kt` is a ~15-line hand-rolled `JUnitCore` runner — the natural
real-world equivalent is `./gradlew test` in an actual Gradle module (which
is how you'd structure this at a company actually running a Bazel/Gradle
Kotlin monorepo); it's included here so the tests are runnable with zero
project setup.

---

## What's actually different per problem

### Problem 1 — Idempotent Deploy-Trigger API

| Concern | Go | Kotlin |
|---|---|---|
| Atomic insert-if-absent | `sync.Mutex` guarding a plain `map` check-and-set | `ConcurrentHashMap.putIfAbsent()` — a single atomic JVM operation, no explicit lock needed for this step |
| Per-key "I own this deploy" lock | not needed — the mutex above already serializes the whole critical section | `ReentrantLock` per key, lazily created via `computeIfAbsent` |
| Durability | hand-rolled write-ahead log: JSON lines appended to a file, `f.Sync()` (fsync) before returning, replayed on startup | same pattern: pipe-delimited lines (no JSON library dependency needed), `RandomAccessFile` + `FileDescriptor.sync()`, replayed on startup |
| Race demonstration | `go test -race` — Go's built-in race detector independently flags the exact bug (a real data-race crash report), in addition to the assertion-based test | JUnit test using a `CyclicBarrier` to force all goroutine-equivalents (threads) to fire at once, then assert `calls > 1` |

**Worth knowing about `sqlite3`/Postgres not being available here:** this
sandbox has no network access to a Go module proxy or a JVM package
repository for a real DB driver, so both ports use a hand-rolled
write-ahead-log-plus-replay for durability instead of an actual SQL
`UNIQUE` constraint (which is what the Python version uses via `sqlite3`,
and what you'd use in production on either platform). The *durability
property being demonstrated* — survive a restart without losing dedup
history or re-running a completed deploy — is identical; only the storage
mechanism differs. Worth saying this out loud in an interview rather than
letting it look like an oversight: "I'd use Postgres with a unique
constraint in production; here's the same guarantee built from `fsync` and
a replay log to keep the example dependency-free."

**A real bug the Go race detector caught that Python's version couldn't
show as directly:** running `go test ./p1/... -race` on the brute-force
version produces an actual `WARNING: DATA RACE` report pointing at the
exact `map` write in `Trigger()` — this is a strictly stronger
demonstration than the assertion-based "deploy fn got called more than
once" test, because it's Go's own tooling independently confirming the bug
exists at the memory-access level, not just at the observable-behavior
level. Worth running both ways (`go test ./p1/...` and `go test ./p1/...
-race`) in a real interview if you get the chance — leading with "let's
also just run the race detector on this" is a strong signal.

### Problem 2 — Rate-Limited Rollback Endpoint

| Concern | Go | Kotlin |
|---|---|---|
| Sliding-window deque | `container/list.List` (doubly linked list) | `java.util.ArrayDeque` |
| Per-service lock | `sync.Mutex` inside a `serviceWindow` struct, looked up via a guarded map | `ReentrantLock` inside a private `ServiceWindow` class, looked up via `ConcurrentHashMap.computeIfAbsent` |
| Forcing concurrent requests to race in tests | unbuffered `chan struct{}` closed to release all goroutines simultaneously (the idiomatic Go "barrier") | `java.util.concurrent.CyclicBarrier` |
| Audit log | mutex-guarded slice | `CopyOnWriteArrayList` (a reasonable real choice here: audit-log reads are rare relative to writes in this shape, and CoW gives lock-free reads) |

Both ports passed the "20 concurrent requests, exactly 3 accepted" test on
every run, including Kotlin's under normal execution and Go's under `-race`.

### Problem 3 — Canary Analysis Evaluator

Nearly a direct transliteration — this problem is pure algorithm with no
concurrency or I/O, so the three languages differ only in deque syntax
(`collections.deque` / `container/list.List` / `java.util.ArrayDeque`) and
in how the differential fuzz test is written (`random` / `math/rand` /
`kotlin.random.Random`, all seeded for reproducibility). All three
differential tests run the same shape of check: 500 randomized points
through both implementations side by side, asserting identical verdicts at
every step.

One genuine Kotlin-specific bug worth mentioning because it's a good
"gotcha to know about the tooling" story: the first version of the Kotlin
test suite used `ev::class.simpleName` (Kotlin reflection) inside assertion
messages, which threw `Kotlin reflection implementation is not found at
runtime` in this apt-installed Kotlin 1.3 toolchain (no `kotlin-reflect.jar`
on the classpath) — 8 of 9 tests failed on that alone, not on any actual
logic bug. Fixed by swapping to `ev.javaClass.simpleName` (plain JVM
reflection, no extra dependency). Worth knowing: `::class` in Kotlin pulls
in `kotlin-reflect` as a real runtime dependency in some toolchains/setups;
`.javaClass` avoids that entirely when you just need a class name for a
log/assertion message.

### Problem 4 — Build-Cache Key Generator

The cleanest of the five to port, since it's pure computation (hashing +
recursion + memoization) with no concurrency:

| Concern | Go | Kotlin |
|---|---|---|
| Hashing | `crypto/sha256` + `encoding/hex` | `java.security.MessageDigest.getInstance("SHA-256")` |
| Cycle detection | `map[string]bool` "currently visiting" set, threaded through recursive calls | `Set<String>` "currently visiting" set, same pattern |
| Memoization | plain `map[string]string` cache | plain `HashMap<String, String>` cache |

The exponential-vs-linear call-count gap (98,303 vs. 32 at diamond depth
16) came out **identical** across Python, Go, and Kotlin, which is exactly
what you'd expect — it's a property of the algorithm and the graph shape,
not the runtime.

### Problem 5 — Change-Failure-Rate Calculator

| Concern | Go | Kotlin |
|---|---|---|
| Per-service partitioned storage | `map[string]*serviceCounters` | `HashMap<String, ServiceCounters>` |
| O(1) event lookup for late-arriving outcomes | `map[int]*list.Element` (pointer straight to the deque node) | `HashMap<Int, DeployEvent>` (direct object reference, since `DeployEvent` fields are mutable `var`s) |
| Locking | single `sync.Mutex` guarding all mutating operations (simpler than per-service locking here, since insert/update/evict all touch shared cross-service maps) | single `ReentrantLock`, same reasoning |
| Concurrent-insert test | goroutines + closed channel barrier, asserting the final CFR is exactly 0.5 after 20 concurrent inserts | threads + `CyclicBarrier`, same assertion |

**One real Kotlin compiler error worth knowing about**, because it's a
sharp edge specific to the language: this
```kotlin
if (wasResolved && !isResolved) sc.resolved--
else if (!wasResolved && isResolved) sc.resolved++
if (wasFailure && !nowFailure) sc.failures--
else if (!wasFailure && nowFailure) sc.failures++
```
as the **last statements in a lambda block** failed to compile with `'if'
must have both main and 'else' branches if used as an expression`. Kotlin
treats the final expression of a lambda as its return value, and tries to
infer a type for that trailing `if/else if` (which has no exhaustive
`else`) as though it were being used as an expression, not a statement.
The fix was a one-line `Unit` appended after the four `if` statements, to
make it unambiguous that the block's return value doesn't come from that
`if` chain. This is a real, easy-to-hit Kotlin gotcha any time a
non-exhaustive `if/else if` (no final plain `else`) happens to land as the
last lines of a lambda — worth having a ready explanation for if it comes
up in a live Kotlin round, since it looks like a strange error message the
first time you see it.

---

## Cross-language takeaway worth stating in the interview

The same three structural fixes show up in every language, just wearing
different syntax:
1. **Replace check-then-act with a single atomic primitive** — SQL's
   `INSERT ... ON CONFLICT DO NOTHING` in Python, a `sync.Mutex`-guarded
   critical section in Go, `ConcurrentHashMap.putIfAbsent` in Kotlin.
2. **Replace O(n) rebuild-from-scratch with an incrementally-maintained
   window** — `collections.deque` / `container/list.List` / `ArrayDeque`,
   identical amortized-O(1) eviction logic in all three.
3. **Replace unmemoized recursion with memoized DAG traversal** — a plain
   dict/map cache in all three, turning exponential into linear.

Being able to say "this is the same fix, and here's how it looks in
whatever language you want me to write it in" is a stronger signal than
being fluent in any one of the three — and it's literally true here, since
the algorithmic logic in all three ports is line-for-line the same shape.
