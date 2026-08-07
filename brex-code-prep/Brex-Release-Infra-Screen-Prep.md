# Brex Technical Screen — 5 Predicted Problems
### Senior Software Engineer, Release Infrastructure (Seattle) · 1-hour CodeSignal screen

---

## 1. What the evidence actually says

I pulled recent candidate reports and cross-referenced them against the recruiter FAQ and the Release Infra job description.
Five things are consistent enough to plan around.

**The format is "Craft," not LeetCode.** Brex's coding rounds start from a small working problem and
layer two or three extensions onto it.
The FAQ's "will NOT try to trick you" line is literal —
    the difficulty is in the extensions, not in spotting a hidden trick.
Candidate reports describe building a small back-end service from scratch (a credit-limit checker, a transaction parser)
in your language of choice, and the onsite rounds are described the same way:
    a base problem plus additional features,
    graded on clean code,
    data modeling
    edge-case handling.

**Scheduling with dependencies is the single most-reported screen topic.** Two separate 2025 phone-screen reports on
1Point3Acres describe exactly this —
one titled around a job scheduler with dependency challenges,
another an OOD task scheduler challenge.
For a Release Infra role this is nearly a lock:
    the build/deploy dependency graph *is* the team's domain.

**Interviewers pivot to idempotency and partial failure.** Multiple sources note that Brex interviewers frequently
interrupt to ask how you'd prevent something from being processed twice,
or what happens on a mid-flight failure. This is the fintech reflex applied to any problem. Expect it as a follow-up
even on a problem that didn't start there.

**Named problem archetypes** that recur across question banks: a `canPurchase`-style rule/eligibility function,
moving-average aggregation over an event stream,
and — called out explicitly as the mapped problem types — *scheduling, state machines, and data parsing*.

**Release Infra scopes the domain.** The JD says you'll own release, observability, and incident processes,
and explicitly names deployment frequency, change failure rate, and MTTR as metrics you'd define and instrument.
Brex's stack is a Bazel monorepo, Kotlin-first, GitOps/K8s. A good interviewer skins the generic problem in that domain
— so expect deploys, pipelines, and incidents rather than credit cards, even though the underlying algorithm is the same.

**One caveat on sourcing:** Glassdoor and 1Point3Acres entries are self-reported, skew toward product-engineering loops,
and none are specifically tagged Release Infra. Treat these as *archetypes with high hit probability*,
not as a leaked question list. The value is that all five share a core — model the domain cleanly, then extend —
so preparing them transfers even if the exact prompt differs.

---

## 2. The five problems

Each has a runnable Kotlin file alongside this doc. All five compile under Kotlin 2.1 and self-verify with inline assertions:

```bash
kotlinc P1DeployScheduler.kt -include-runtime -d p1.jar && java -jar p1.jar
```

| # | Problem                         | Core skill                    | Confidence            |
|---|---------------------------------|-------------------------------|-----------------------|
| 1 | Deploy dependency scheduler     | Graphs / topological sort     | **Highest**           |
| 2 | Release log → DORA metrics      | Parsing + aggregation         | High                  |
| 3 | Deploy gate / rollout policy    | OOD + open-closed extension   | High                  |
| 4 | Idempotent event processor      | Hash maps + failure semantics | High as a *follow-up* |
| 5 | Release lifecycle state machine | FSM / data modeling           | Medium-high           |

---

### P1 — Deploy Dependency Scheduler
**File:** `P1DeployScheduler.kt`

> *"We deploy a few hundred services out of a monorepo. Each service declares which other services must be deployed before it. Given that list, produce a valid deploy order."*

**Why this one:** two independent 2025 phone-screen reports, and it's the closest thing to the team's actual daily problem.

**Clarify before coding — say these out loud:**
- Can a service depend on something that isn't in the input list? *(Config bug — I'll surface it distinctly from a cycle.)*
- Should I return one serial order, or groups that can run in parallel? *(Ask; then build waves, since it's strictly more informative and flattens for free.)*
- Are cycles possible, and do you want the cycle named or just rejected?
- Should output order be deterministic for equal-priority jobs? *(Sort within a wave — makes it testable.)*

**Approach:** Kahn's algorithm, O(V + E). Track in-degree, drain the zero-in-degree frontier one wave at a time.
If you emit fewer nodes than you were given, you have a cycle — then run a DFS restricted to the residual nodes to name it.

**The extensions they'll add, in the order they usually come:**
1. **Parallel waves** — group instead of serialize. Nearly free if you built Kahn's by frontier.
2. **Cycle reporting** — "invalid graph" is a weak answer; `a → c → b → a` is the answer. Colored DFS (unvisited / on-stack / done) over the residual set.
3. **K workers + durations** — event-driven simulation with two priority queues (ready-at, finish-at). Makespan with unlimited workers is the critical path; with one worker it's the sum. Any K sits between.
4. **Possible 4th:** given a changed file, which services need rebuilding? (Reverse-reachability from the changed node.)

**Traps:** using recursion depth proportional to graph size without mentioning it; conflating "missing dependency" with "cycle"; non-deterministic ordering that makes your own test flaky.

---

### P2 — Release Event Log → DORA Metrics
**File:** `P2ReleaseMetrics.kt`

> *"Here's a raw event log from our deploy pipeline. Compute per-service deployment frequency, change failure rate, and mean time to restore."*

**Why this one:** the parsing/aggregation archetype ("transaction parser," "moving average of transaction volumes") mapped onto the exact three metrics the JD names. If your interviewer wants to test data handling rather than graphs, this is the shape it takes.

**Clarify:**
- What's the line format, and is it guaranteed well-formed? *(Assume not — decide whether to skip-and-report or fail fast. Skip-and-report is almost always the right call for a pipeline; say why.)*
- Are events guaranteed in timestamp order? *(No. Producers batch and retry. Sort once at parse.)*
- How is an incident attributed to a deploy — does the incident carry a deploy ID, or do we blame the most recent prior deploy?
- What if an incident is never resolved? Does it count against MTTR, or get reported separately?
- Over what window is frequency normalized?

**Approach:** parse into typed events with a rejected-lines report, sort once, group by service, then three independent reducers. The interesting judgment calls are all in the definitions, not the code — say them out loud as you make them:
- CFR counts **distinct** blamed deploys, so two incidents from one bad deploy don't double-count.
- MTTR excludes unresolved incidents but reports their count separately, because silently dropping them makes the metric look better as reliability gets worse.

**Extensions:** rolling/windowed metrics; a moving average over a trailing N deploys; p50/p95 restore time instead of the mean (and why the mean is the worse metric here).

**Traps:** parsing with a fragile regex over a `split`-and-validate loop; throwing on the first bad line; not stating your incident-attribution rule before implementing it.

---

### P3 — Deploy Gate / Rollout Policy Engine
**File:** `P3DeployGate.kt`

> *"Write `canDeploy(request)`. A deploy is allowed if all required CI checks are green."* …then three more requirements arrive.

**Why this one:** this is the `canPurchase`-style rule function, which is Brex's signature Craft shape. The base is trivially easy on purpose — the entire signal is in how your first design absorbs requirements 2, 3, and 4.

**The trap is the whole point.** If you write the base as an if-chain inside one function, each extension makes it worse and you'll be visibly refactoring under time pressure. If you write it as a list of `Rule` objects behind a small interface, every extension is a new class and the engine never changes. Set that up in the first two minutes and *say why*: "I'll model each policy as its own rule so we can add more without touching the evaluator."

**Extensions to expect:**
1. **Change-freeze windows** (no prod deploys Friday evening → Monday morning).
2. **Explain the denial** — return every violation, not just the first. Ops needs to fix all of them in one pass, not play whack-a-mole.
3. **Break-glass override** — a label that downgrades a hard deny to "needs approval." Then the real question: *which rules must break-glass never override?* Red security scans, in my answer. Have an opinion here.
4. **Promotion ordering** — the artifact must be live in staging before production.

**Traps:** returning a bare boolean (you can't explain a denial); short-circuiting on the first violation; letting break-glass override everything with no exceptions.

---

### P4 — Idempotent Event Processor
**File:** `P4IdempotentProcessor.kt`

> *"Our CI provider delivers webhooks at-least-once. Make sure each event is processed exactly once."*

**Why this one:** less likely as the opening prompt, near-certain as a follow-up. Reports consistently describe interviewers pivoting to "how do you stop this being processed twice?" and "what if the DB dies mid-flight?" Prepare it as a *pivot*, not a standalone.

**Clarify:**
- Where does the idempotency key come from — the provider, or do we derive it?
- If the handler throws, is that "processed"? *(No — and this is the crux.)*
- Do duplicates arrive concurrently, or only sequentially?
- How long must the dedupe record live? *(Longer than the provider's retry window.)*

**The four things that separate a good answer from a shallow one:**
1. **Failures are not cached.** A naive `if (seen) return` marks the key on entry and permanently swallows a transient failure. Commit the key only after the body succeeds.
2. **Store the result, not just the flag.** A duplicate should return the same answer the first call returned.
3. **TTL with an injected clock.** Unbounded dedupe maps are a memory leak. An injected clock is also the only way to test expiry without sleeping.
4. **Concurrency.** Two threads hitting the same key must serialize. `ConcurrentHashMap` alone doesn't give you this — check-then-act still races. Per-key lock, double-checked under the lock. The included file proves it with 16 threads on one key: 1 execution, 15 duplicates.

**Extension:** out-of-order versioning — a stale update for a key should be dropped, not applied over the newer one.

**If asked to make it durable:** the in-memory map becomes a unique constraint on `(idempotency_key)` in Postgres, insert-then-work inside the same transaction as the side effect, or an outbox if the side effect is external. Say this even if you don't code it — it's the answer they're fishing for.

---

### P5 — Release Lifecycle State Machine
**File:** `P5ReleaseStateMachine.kt`

> *"Model a release going from pending through canary to fully live, with rollback. Reject illegal transitions."*

**Why this one:** state machines are named explicitly as a mapped archetype, and this is the natural data-modeling question for a team that owns releases and incidents. If your screen is a "data modeling" question rather than a DS&A one, this is very likely its shape.

**Clarify:**
- Which states, and is the list open to extension?
- Is rollback allowed from every state or only some?
- Are there time-based conditions (canary bake time)?
- Do you need history, or just the current state? *(Assume history — this is a compliance-adjacent system at a fintech.)*

**Approach:** a transition **table** (`Map<(State, Trigger), State>`), not a `when` block. Guards that the table can't express (bake time, traffic percentage) go in a separate `guard()` step. Every transition appends to an immutable audit record with actor, timestamp, and reason.

**Extensions:** minimum canary bake before ramping; staged traffic percentages (5 → 25 → 50 → 100) with the bake clock resetting each step; rollback from any live-ish state; queries like "how long was this in RAMPING?"

**Traps:** a nested `when` that grows quadratically; `LIVE` modeled as fully terminal so rollback becomes impossible; no audit trail; time-based guards using `Instant.now()` directly, which makes the whole thing untestable.

---

## 3. Language call

**Use Kotlin.** Brex's backend is Kotlin-first with a Bazel monorepo, and the FAQ explicitly blesses static languages. Your Gradle/Bazel/Kotlin background is the reason you're in this loop — writing Python would quietly discard your strongest signal. Kotlin's data classes, sealed interfaces, and `when` exhaustiveness also make you *look* fast at exactly the modeling this role tests.

One hedge: if the problem turns out to be heavy string munging with no domain modeling, Python is faster to type. Decide in the first two minutes and don't switch after.

---

## 4. The 50 minutes

The FAQ is unusually explicit that communication is graded, and Glassdoor reports back it up — candidates fail this loop on process, not on correctness.

**0–5 min — Clarify.** Ask 3–4 questions before typing anything. State every assumption aloud. "I'm assuming dependencies are declared by name and the list is complete — tell me if that's wrong."

**5–10 min — Model out loud.** Name your types before implementing them. "I'll use a `Job` data class and a `sealed interface` for the result so invalid inputs are a value, not an exception." This is the single highest-leverage minute of the interview for a Release Infra hire.

**10–35 min — Build the base, make it run.** The FAQ says a running solution beats a perfect one. Get something executing early, then improve.

**35–50 min — Take the extensions.** This is where the grade is decided. If your base was modeled well, extensions are additive and you'll look calm.

**Debugging.** Some loops include a round where you must find *all* the bugs — one report notes that missing one can sink the loop. Whatever the round, narrate a deterministic strategy: reproduce → isolate with a minimal case → hypothesis → verify. Don't guess-and-check silently.

**One thing to weave in:** you've *built* this system. The release branching tool that went from three weeks to two hours, the GraphQL rewrite that cut per-run lookups from an hour to seconds across hundreds of projects, the CMake-to-Bazel migration. Don't lecture — but when the dependency-graph problem appears, "this is close to something I've shipped" is a strong, honest note to strike.

---

## 5. Questions to ask them

The FAQ tells you to prepare these. Release-Infra-specific ones land better than generic ones:

- What does the release pipeline look like end to end today, and where's the biggest source of friction?
- How does the team measure itself — are DORA metrics instrumented and acted on, or aspirational?
- How mature is the Bazel setup? Remote caching, remote execution, or mostly local?
- Release Infra owns observability and incident response too. How's that split with Cloud Infra and Application Infra?
- What's changed for Infrastructure since the Capital One acquisition — budget, compliance requirements, roadmap?
- What would you want this person to have shipped by month six?

---

## Sources

Recent candidate reports on Glassdoor and 1Point3Acres (2024–2026), aggregated process write-ups on TechPrep and Interview Query, the Brex Release Infra and Application Infrastructure job postings, and the recruiter FAQ you were sent. Self-reported interview data is noisy — the archetypes are reliable, the exact prompts are not.
