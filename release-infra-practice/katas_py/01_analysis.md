# Base Problem Analysis

These are read as **systems-flavored** interview problems, not textbook LeetCode
puzzles. You're expected to:
- call/consume an API-shaped payload (a list of dicts / JSON)
- handle real date-time parsing correctly
- produce a correct, well-typed result

with a working algorithm underneath. Python is used throughout since it minimizes ceremony for JSON + datetime handling
(the exact trade-off called out for this interview format vs. Java).

Reference implementations for all three live in `02_base_solutions.py` and are
runnable directly (`python3 02_base_solutions.py`) (each has embedded assertions
so you get pass/fail feedback).

---

## 1. `first_duplicate_deployment`

### **Read of the problem:**

Given a stream/list of deployment records (e.g. build IDs, commit SHAs, or `{"deployment_id": ..., "timestamp": ...}`
objects pulled from a CI/CD API), find the **first ID that has already been seen before** aka
the first element in scan order whose duplicate you already passed.
(Distinct from "the duplicate value that first *completes* a pair" — same answer here since we scan left to right and stop at the first repeat.)

**Algorithm:** single pass, hash set of seen IDs.

```
seen = set()
for record in stream:
    if record.id in seen:
        return record.id  # or record, or index
    seen.add(record.id)
return None
```

- **Time:** O(n) — one pass, O(1) average hash set membership check/insert.
- **Space:** O(n) worst case (all-unique-until-the-end) for the `seen` set.

**Trade-offs / optimizations:**
- **Space-for-time is basically unavoidable** here if input arrives as a true stream (can't re-scan).
- If the full list is materialized in memory up front and IDs are small bounded integers (e.g., deployment IDs in a known range)
- you can trade the hash set for a **bitset** (`bytearray` or bit-packed int), cutting constant factors and memory ~8-64x
  - (but this only helps when the ID domain is dense/bounded, which is rare for UUID-style deployment IDs).
- **Sort-based alternative:**
  - sort then scan adjacent pairs — O(n log n) time
  - O(1) extra space (or O(n) if you can't sort in place / need to preserve original order for the *first* occurrence).
  - This is strictly worse on time and also **destroys temporal order**, so it only answers "does a duplicate exist" or "which ID is duplicated," not "which one appeared first in the stream"
    - (a real regression for this problem's actual intent).
  - Don't use it if "first" matters.
- **API angle:**
  - if `stream` is actually paginated API results, the real engineering trade-off is whether you can dedupe *while paginating*
    - (stream processing, O(1) extra pages in flight) vs. materializing every page first (simpler code, more memory, higher latency-to-first-possible-answer).
    - For an interview, mention both and default to processing-as-you-fetch.
- **Concurrency:**
- if deployments can be reported out of order (e.g. from
  multiple regions)
  - "first" becomes ambiguous
    - clarify whether "first" means by stream-arrival order or by a `timestamp` field.
  - If it's by timestamp, you must sort first (see problem 2's pattern) before doing the linear scan
    - changing the complexity to O(n log n).

---

## 2. `parse_and_sort_events`

**Read of the problem:**
Given raw event data (JSON from a REST response, or a list of dict-like records),
with a timestamp field in some format (ISO 8601, Unix epoch, RFC 2822,
                                        or inconsistent mixes thereof — a classic "gotcha" in this style of question)
parse each into a structured/typed form and sort by time.

**Algorithm:**
1. Parse each raw record into a typed object (validate/normalize datetime — this is the actual hard part, not the sort).
2. Sort using the parsed datetime as key.

```
parsed = [parse_event(r) for r in raw_events]
parsed.sort(key=lambda e: e.timestamp)
```

- **Time:** `O(n)` to parse + `O(n log n)` to sort = **O(n log n)** dominant.
- **Space:** `O(n)` for the parsed structures (unavoidable — you need the parsed output anyway);
  - Python's Timsort is `O(n)` auxiliary space worst case, `O(1)` amortized for nearly-sorted input.

**Trade-offs / optimizations:**
- **Timezone correctness is the real risk**, not big-O.
  - `datetime.fromisoformat` doesn't handle a trailing `Z` before Python 3.11
  - mixing naive and timezone-aware datetimes raises `TypeError` on comparison
- The safe pattern:
  - normalize everything to UTC-aware `datetime` (or epoch `float`) immediately at parse time
  - never compare naive-to-aware.
- **Streaming/incremental sort:**
  - if events arrive continuously (e.g. paginated API results, or a live feed) and you need a running sorted view
    - a **heap** (`heapq`) gives O(log n) insert instead of full O(n log n) re-sorts
      - use this when the problem is really "maintain sorted order over time," not "sort a fixed batch."
- **If already nearly sorted** (e.g., each API page is internally sorted, you just need to merge k sorted pages)
  - use a **k-way merge** (`heapq.merge`)
    - `O(n log k)` instead of `O(n log n)`,
    - `O(k)` space instead of `O(n)` since you don't need to materialize the merged list if you can stream it.
- **Stable sort matters** if there's a tiebreaker (e.g., same timestamp, preserve original arrival order)
  - Python's `sort`/`sorted` are stable by default, which is a free correctness win worth calling out explicitly in an
    interview.
- **Parsing failures:**
- decide and state your policy
  - skip malformed records
  - raise
  - quarantine them into an errors list
- this is the kind of judgment call this interview format is testing more than algorithmic cleverness.

---

## 3. `max_profit_schedules`

**Read of the problem:** Classic **weighted interval scheduling**: given a set
of jobs/schedules, each with a `start`, `end`, and `profit`, choose a subset of
**non-overlapping** intervals that maximizes total profit.

**Algorithm:** sort by end time, then DP with binary search for the latest
non-conflicting predecessor (`p(i)`).

```
jobs.sort(key=lambda j: j.end)
dp = [0] * (n + 1)
for i in 1..n:
    p = last index j < i where jobs[j].end <= jobs[i-1].start   # binary search
    dp[i] = max(dp[i-1], dp[p] + jobs[i-1].profit)
return dp[n]
```

- **Time:** O(n log n) — sort dominates; the DP loop with binary search is
  O(n log n) as well (n lookups, each O(log n)).
- **Space:** O(n) for the DP table (can be reduced to O(1) extra if you only
  need the *value*, not the reconstructed job list — but reconstructing which
  jobs were chosen needs the full table or a parallel "choice" array).

**Trade-offs / optimizations:**
- **Greedy-by-profit or greedy-by-duration does NOT work** — this is the
  classic trap that distinguishes this problem from plain (unweighted)
  interval scheduling, where greedy-by-end-time is optimal. Calling this out
  explicitly is a strong signal in interview: "if all profits were equal this
  would reduce to activity selection and greedy would be optimal; because
  profits vary, we need DP."
- **`p(i)` via binary search vs. two-pointer:** since jobs are sorted by end
  time, and you're querying "latest end ≤ this start," you *can* use a
  two-pointer/merge-style scan instead of binary search if you also sort a
  second array by start — gets you O(n) after the initial O(n log n) sorts,
  same asymptotic total but fewer log-factor operations in practice.
- **Memory-constrained variant:** if n is huge (streaming jobs from an API,
  can't hold all in memory), full DP isn't directly streamable since `dp[i]`
  can depend on any earlier `dp[p]`. A common real-world compromise: bucket
  jobs into non-overlapping time windows first (if the domain allows, e.g. by
  day) and solve each window independently, trading global optimality for
  bounded memory.
- **Reconstruction:** if the interview asks for *which* jobs (not just the max
  profit number), keep a parallel `choice[i]` boolean/array during the DP and
  walk it backward — O(n) extra space, O(n) extra time, worth mentioning
  proactively since "just the number" is a weaker answer.
- **API angle:** if `jobs` come from paginated endpoints, you generally must
  materialize the full set before running DP (global optimization needs global
  visibility) — unlike problems 1 and 2, this one does **not** parallelize
  well with a streaming/incremental approach without approximation.

---

## Cross-cutting patterns worth naming out loud in the interview

| Problem | Core technique | Why sort (if any) | Streamable? |
|---|---|---|---|
| first_duplicate_deployment | Hash set, single pass | No | Yes |
| parse_and_sort_events | Parse + sort (or heap/k-way merge) | Yes — establishes order | Partially (heap for running order) |
| max_profit_schedules | Sort + DP + binary search | Yes — enables DP transition | No (needs global view) |

Naming this table out loud ("this is a dedup pass," "this is establishing order
before I can reason about anything," "this needs global visibility so I can't
stream it") is exactly the kind of signal that separates "wrote working code"
from "understood the problem" in a systems-style interview.
