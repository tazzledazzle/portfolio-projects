# 20 Kata Variations

Same three algorithmic cores as the base problems, wrapped in the kind of
"small working piece of a real system" framing this interview format favors:
API-shaped input, JSON, dates/timezones, and a concrete business scenario.
Working solutions for all 20 are in `04_kata_solutions.py`, grouped by
category and runnable directly.

Target: read the statement, name the pattern out loud, state target time/space
complexity *before* coding, then implement.

---

## Category A — Dedup / "first repeat" pattern (hash set, single pass)
*(variations on `first_duplicate_deployment`)*

**A1. `first_replayed_request`**
Your payments API logs `{"request_id": str, "endpoint": str, "ts": str}` for
every inbound call. A buggy retry client sometimes replays a request with the
same `request_id`. Given the log in arrival order, return the `request_id` of
the first replayed request, or `None`. Target: O(n) time, O(n) space.

**A2. `first_ip_over_threshold`**
Given a stream of `{"ip": str, "ts": str}` login attempts, return the first IP
address that appears **more than `k` times within any rolling 60-second
window** (a lightweight brute-force intrusion signal). Target: discuss why
this is no longer a plain hash-set problem — it needs a sliding window per IP.

**A3. `first_duplicate_idempotency_key`**
A payment-processing queue consumes `{"idempotency_key": str, "amount_cents":
int}` messages. Return the first message whose `idempotency_key` has already
been consumed (indicating a duplicate charge risk), *and* verify the `amount`
matches the first occurrence — flag as `"conflict"` if it doesn't. Target:
extend the hash set to a hash map storing the first-seen payload.

**A4. `first_reused_session_token`**
Given `{"token": str, "region": str, "ts": str}` login events, detect the
first token used from a **second, different region** — a basic session-hijack
signal. Target: hash map from token → first-seen region, O(n).

**A5. `first_repeated_device_fingerprint`**
Given signup attempts `{"device_fingerprint": str, "email": str}`, return the
first fingerprint reused across two *different* email addresses (possible
fraud ring), ignoring legitimate reuse by the same email. Target: map
fingerprint → set of emails seen, flag on second distinct email.

**A6. `first_cycle_in_redirect_chain`**
Given a webhook redirect chain as an ordered list of URLs visited, return the
first URL that has already appeared earlier in the same chain (infinite-loop
detection before you follow one more hop). Target: O(n) hash set, but discuss
why this differs from cycle detection in a *graph* (here it's a fixed
sequence, not something you're traversing live) — and when you'd need
Floyd's cycle detection instead (traversing a live `next` pointer structure
with unbounded/streaming length and O(1) space constraint).

**A7. `first_stale_cache_key_collision`**
Given a stream of cache-write events `{"key": str, "version": int, "ts":
str}`, return the first `key` written with a **lower `version`** than a
version already recorded for that key (out-of-order write / stale write
detection). Target: hash map from key → max version seen so far.

---

## Category B — Parse + sort / establish order pattern
*(variations on `parse_and_sort_events`)*

**B1. `merge_paginated_commits`**
Given 3 "pages" of a GitHub-shaped commits API response (each a list of
`{"sha": str, "committed_at": str}`, each page already sorted by time within
itself but pages arrive in arbitrary order and may overlap at boundaries),
produce a single deduplicated list sorted ascending by `committed_at`. Target:
discuss k-way merge (`heapq.merge`) vs. full re-sort, and where dedup fits in.

**B2. `normalize_mixed_format_transactions`**
Given a CSV-like list of transaction dicts where `timestamp` may be ISO 8601,
Unix epoch (int or str), or `"MM/DD/YYYY HH:MM"` US-format — parse each into a
UTC-aware `datetime`, drop (into a separate `errors` list) anything
unparseable, then sort the valid ones ascending. Target: correctness of the
parsing dispatcher matters more than the sort.

**B3. `flatten_and_order_thread_replies`**
Given a nested Slack-shaped payload: top-level messages each with a `replies`
list (each reply itself has a timestamp), flatten into a single list of all
messages+replies and sort by `ts`, but **break ties by thread hierarchy**
(parent message must sort before any of its replies even if timestamps are
identical due to clock resolution). Target: composite sort key.

**B4. `parse_and_rank_log_severity`**
Given raw log lines like `"[2026-01-01T00:00:03Z] ERROR: disk full"`, parse
into `{level, timestamp, message}`, then sort primarily by severity rank
(`CRITICAL > ERROR > WARN > INFO`) and secondarily by timestamp ascending
within the same severity. Target: regex/parsing correctness + composite sort
key with a custom rank mapping.

**B5. `convert_and_sort_multi_currency_ledger`**
Given transactions `{"amount": float, "currency": str, "ts": str}` and a
separate `rates: dict[str, float]` (currency → USD rate) fetched from an FX
API, convert every amount to USD and return transactions sorted by USD amount
descending. Target: handle a missing-rate currency gracefully (skip + report,
don't crash).

**B6. `resolve_recurring_calendar_events`**
Given events with an optional `"recurrence": {"freq": "daily"|"weekly",
"count": int}` field and a base `start` time, expand each recurring event into
its concrete occurrences within a given date range, then return the full
occurrence list (recurring + one-off) sorted by start time. Target: generation
+ sort, discuss memory blow-up risk if `count` or the range is huge (generator
vs. materialized list).

**B7. `dedupe_and_sort_api_pages_by_etag`**
Given paginated API results where retries may have fetched the same page
twice (identical `id` + `etag`), reduce to unique records (keep the version
with the latest `etag` if they differ) and sort by a `rank` field ascending,
with `id` as a stable tiebreaker. Target: dedup-with-conflict-resolution
folded into the sort pipeline.

---

## Category C — Weighted interval scheduling / max-profit pattern
*(variations on `max_profit_schedules`)*

**C1. `max_value_spot_instance_reservations`**
Given cloud spot-instance bid requests `{"start": int, "end": int,
"bid_value": int}` on a single machine, select a non-overlapping subset that
maximizes total accepted bid value. Target: identical shape to the base
problem — confirm the mapping explicitly, then implement.

**C2. `max_revenue_ad_slot_booking`**
Given ad campaign requests `{"window_start": int, "window_end": int,
"revenue": int}` for a single ad slot, choose non-overlapping bookings
maximizing total revenue, but this time also return **which campaigns were
rejected** (for a "sorry, that slot's taken" notification). Target: same DP,
plus a set-difference reconstruction step.

**C3. `max_priority_room_bookings`**
Given meeting requests `{"start": int, "end": int, "priority": int}` for a
single conference room, maximize total priority score of accepted meetings.
Add a twist: two meetings with `end == start` are considered **overlapping**
here (no back-to-back allowed, unlike the base problem) — discuss how a
single inequality (`<` vs `<=`) in the binary search changes correctness.

**C4. `max_payout_freelance_gig_schedule`**
Given gig offers `{"start": int, "end": int, "payout": int}` for one
freelancer who can only work one gig at a time, maximize total payout, and
additionally support a **query**: "what's the max payout achievable using
only gigs that end by time T?" for several different T values efficiently
(reuse the DP table rather than recomputing). Target: precompute once, answer
each query in O(log n) via binary search into the DP array.

**C5. `max_stipend_oncall_shift_bidding`**
Given on-call shift bids `{"start": int, "end": int, "stipend": int,
"engineer_id": str}`, select non-overlapping shifts maximizing total stipend
paid out, but **cap total shifts assigned to any single engineer at 2** — a
resource-constraint twist that breaks the simple 1-D DP. Target: recognize
this needs a 2-D DP (index, count-per-engineer) or a different formulation,
and at minimum articulate why the greedy/simple-DP approach no longer applies
even if a full efficient solution isn't reached.

**C6. `max_profit_with_setup_time`**
Same as the base problem, but every job requires a fixed `setup_minutes` of
buffer *before* it can start after the previous job ends — i.e., job `i` can
only follow job `j` if `job[i].start >= job[j].end + setup_minutes`. Maximize
total profit. Target: show the one-line change to the binary-search bound
(`job.start - setup_minutes` instead of `job.start`) and confirm the rest of
the DP is unaffected.

---

## Suggested practice order

1. Do **A1–A3** cold (should feel close to identical to the base problem —
   build speed).
2. Do **B2 and B4** — these are where "systems" correctness (parsing,
   timezones, tie-breaking) matters more than algorithmic novelty.
3. Do **C1** cold, then **C3** and **C6** (small but sharp variations on the
   DP transition — good for catching off-by-one boundary bugs under time
   pressure).
4. Attempt **C5** last — it's intentionally the one that doesn't reduce
   cleanly to the base pattern, to practice recognizing *when* a familiar
   template stops applying.
