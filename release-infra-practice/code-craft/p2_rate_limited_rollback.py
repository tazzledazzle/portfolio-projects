"""
Problem 2: Rate-Limited Rollback Endpoint
==========================================
Accepts rollback requests for a service. Allows at most 3 rollbacks per
service per 10-minute SLIDING window (to prevent flapping — a bad rollback
loop making things worse). Returns 429 + Retry-After when exceeded. Every
ACCEPTED rollback is written to an append-only audit log.

Run:
    python3 -m unittest p2_rate_limited_rollback -v
"""
from __future__ import annotations

import threading
import time
import uuid
from collections import deque
from dataclasses import dataclass
from typing import Optional


WINDOW_SECONDS = 10 * 60
MAX_ROLLBACKS_PER_WINDOW = 3


@dataclass(frozen=True)
class RollbackResult:
    accepted: bool
    retry_after_seconds: Optional[float] = None
    audit_id: Optional[str] = None


class AppendOnlyAuditLog:
    """Durable, append-only. In production this is a file opened with
    O_APPEND / fsync, or an INSERT-only DB table — never mutated, never
    deleted from the request path. Kept as a simple in-memory list here
    with a lock, since the interesting design decisions (what gets
    logged, and WHEN relative to the rate-limit decision) are independent
    of the storage backend."""

    def __init__(self):
        self._lock = threading.Lock()
        self._entries: list[dict] = []

    def append(self, service_name: str, requested_at: float) -> str:
        entry_id = str(uuid.uuid4())
        with self._lock:
            self._entries.append({
                "id": entry_id,
                "service_name": service_name,
                "requested_at": requested_at,
            })
        return entry_id

    def all_entries(self) -> list[dict]:
        with self._lock:
            return list(self._entries)


# ---------------------------------------------------------------------------
# STEP 1 — BRUTE FORCE
# ---------------------------------------------------------------------------
# Reasoning:
#   Keep a plain Python list of timestamps per service. On every request:
#   rebuild the list by filtering out anything older than the window,
#   check the remaining length against the limit, and append if allowed.
#   Correct, but does O(n) work (a full list scan + rebuild) on EVERY
#   single request, where n = number of rollbacks for that service in the
#   last 10 minutes. For a small limit like 3 this is fine, but it's the
#   wrong instinct to generalize: if this were "rate limit deploys to
#   500/hour", you'd be scanning up to 500 entries per request for no
#   reason. It's also not thread-safe: two concurrent requests can both
#   read a list of length 2, both decide "2 < 3, allowed", and both append
#   — letting 4 rollbacks through the "limit of 3" window.
# ---------------------------------------------------------------------------
class BruteForceRateLimiter:
    def __init__(self, audit_log: AppendOnlyAuditLog):
        self._timestamps: dict[str, list[float]] = {}
        self._audit_log = audit_log

    def try_rollback(self, service_name: str, now: Optional[float] = None) -> RollbackResult:
        now = now if now is not None else time.time()
        # --- THE BUG: read, decide, write are three separate,
        #     unsynchronized steps. ---
        history = self._timestamps.get(service_name, [])

        # O(n) rebuild every call.
        window_start = now - WINDOW_SECONDS
        recent = [t for t in history if t > window_start]

        if len(recent) >= MAX_ROLLBACKS_PER_WINDOW:
            oldest_in_window = min(recent)
            retry_after = oldest_in_window + WINDOW_SECONDS - now
            return RollbackResult(accepted=False, retry_after_seconds=max(0.0, retry_after))

        time.sleep(0.005)  # model realistic work between decision and write, widening the race
        recent.append(now)
        self._timestamps[service_name] = recent
        audit_id = self._audit_log.append(service_name, now)
        return RollbackResult(accepted=True, audit_id=audit_id)


# ---------------------------------------------------------------------------
# STEP 2 — OPTIMAL
# ---------------------------------------------------------------------------
# Reasoning:
#   Two independent fixes:
#
#   (a) Efficiency: use a per-service deque as a sliding-window log.
#       New timestamps are always appended in (roughly) increasing time
#       order, so expired entries are always at the LEFT of the deque.
#       Evicting them is O(1) amortized per call (each timestamp is
#       pushed once and popped once, ever) instead of O(n) rebuild.
#       Space is bounded by MAX_ROLLBACKS_PER_WINDOW+1, not by total
#       request volume — this matters a lot once the limit is "500/hour"
#       instead of "3/10min".
#
#   (b) Correctness under concurrency: guard the read-decide-write
#       sequence for a given service with a per-service lock, so the
#       "check length, then append" step is atomic. We use a lock PER
#       SERVICE (not one global lock) so that rollback requests for
#       different services never contend with each other — a flapping
#       payments-service shouldn't slow down rollbacks for an unrelated
#       team's service.
#
#   Audit log ordering: we append to the audit log only AFTER the
#   in-memory decision is committed, but still INSIDE the same critical
#   section — so "accepted" and "logged" can never disagree. If the
#   audit write itself needs to be durable across a crash between
#   "decided accepted" and "logged", that argues for making the sliding
#   window state and the audit log the SAME durable store in production
#   (e.g., both rows in the same DB transaction) — noted as a follow-up,
#   not implemented here, since the prompt's audit log is explicitly a
#   simple append-only log rather than the rate limiter's source of truth.
# ---------------------------------------------------------------------------
class OptimalRateLimiter:
    def __init__(self, audit_log: AppendOnlyAuditLog):
        self._windows: dict[str, deque[float]] = {}
        self._locks: dict[str, threading.Lock] = {}
        self._locks_guard = threading.Lock()
        self._audit_log = audit_log

    def _lock_for(self, service_name: str) -> threading.Lock:
        with self._locks_guard:
            if service_name not in self._locks:
                self._locks[service_name] = threading.Lock()
            return self._locks[service_name]

    def try_rollback(self, service_name: str, now: Optional[float] = None) -> RollbackResult:
        now = now if now is not None else time.time()
        lock = self._lock_for(service_name)
        with lock:
            window = self._windows.setdefault(service_name, deque())

            # Evict expired entries from the left. Amortized O(1): each
            # timestamp is popped at most once across its whole lifetime.
            window_start = now - WINDOW_SECONDS
            while window and window[0] <= window_start:
                window.popleft()

            if len(window) >= MAX_ROLLBACKS_PER_WINDOW:
                oldest = window[0]
                retry_after = max(0.0, oldest + WINDOW_SECONDS - now)
                return RollbackResult(accepted=False, retry_after_seconds=retry_after)

            window.append(now)
            audit_id = self._audit_log.append(service_name, now)
            return RollbackResult(accepted=True, audit_id=audit_id)


# ---------------------------------------------------------------------------
# Tests
# ---------------------------------------------------------------------------
import unittest  # noqa: E402


class TestBruteForceRaceCondition(unittest.TestCase):
    def test_concurrent_requests_can_exceed_the_limit(self):
        log = AppendOnlyAuditLog()
        limiter = BruteForceRateLimiter(log)
        barrier = threading.Barrier(10)
        results: list[RollbackResult] = []
        results_lock = threading.Lock()

        def fire():
            barrier.wait()
            r = limiter.try_rollback("checkout-service")
            with results_lock:
                results.append(r)

        threads = [threading.Thread(target=fire) for _ in range(10)]
        for t in threads:
            t.start()
        for t in threads:
            t.join()

        accepted = sum(1 for r in results if r.accepted)
        self.assertGreater(accepted, MAX_ROLLBACKS_PER_WINDOW,
                            "expected the brute-force limiter to let more than "
                            f"{MAX_ROLLBACKS_PER_WINDOW} through under a race")


class TestOptimalRateLimiter(unittest.TestCase):
    def setUp(self):
        self.log = AppendOnlyAuditLog()
        self.limiter = OptimalRateLimiter(self.log)

    def test_allows_up_to_the_limit(self):
        base = 1_000_000.0
        for i in range(MAX_ROLLBACKS_PER_WINDOW):
            r = self.limiter.try_rollback("svc-a", now=base + i)
            self.assertTrue(r.accepted, f"rollback #{i+1} should be accepted")

    def test_rejects_the_request_over_the_limit_with_retry_after(self):
        base = 1_000_000.0
        for i in range(MAX_ROLLBACKS_PER_WINDOW):
            self.limiter.try_rollback("svc-a", now=base + i)
        r = self.limiter.try_rollback("svc-a", now=base + MAX_ROLLBACKS_PER_WINDOW)
        self.assertFalse(r.accepted)
        self.assertIsNotNone(r.retry_after_seconds)
        self.assertGreater(r.retry_after_seconds, 0)

    def test_window_boundary_exact_expiry(self):
        base = 1_000_000.0
        self.limiter.try_rollback("svc-a", now=base)
        self.limiter.try_rollback("svc-a", now=base + 1)
        self.limiter.try_rollback("svc-a", now=base + 2)
        # Exactly WINDOW_SECONDS later, the first entry (at `base`) must
        # have expired (we use a half-open window: entries at exactly the
        # boundary are excluded), freeing up a slot.
        r = self.limiter.try_rollback("svc-a", now=base + WINDOW_SECONDS)
        self.assertTrue(r.accepted, "the oldest entry should have aged out of the window by now")

    def test_one_second_before_expiry_is_still_rejected(self):
        base = 1_000_000.0
        self.limiter.try_rollback("svc-a", now=base)
        self.limiter.try_rollback("svc-a", now=base + 1)
        self.limiter.try_rollback("svc-a", now=base + 2)
        r = self.limiter.try_rollback("svc-a", now=base + WINDOW_SECONDS - 1)
        self.assertFalse(r.accepted)

    def test_services_have_independent_limits(self):
        base = 1_000_000.0
        for i in range(MAX_ROLLBACKS_PER_WINDOW):
            self.limiter.try_rollback("svc-a", now=base + i)
        # svc-b should be completely unaffected by svc-a hitting its limit.
        r = self.limiter.try_rollback("svc-b", now=base)
        self.assertTrue(r.accepted)

    def test_concurrent_requests_never_exceed_the_limit(self):
        barrier = threading.Barrier(20)
        results: list[RollbackResult] = []
        results_lock = threading.Lock()

        def fire():
            barrier.wait()
            r = self.limiter.try_rollback("svc-concurrent")
            with results_lock:
                results.append(r)

        threads = [threading.Thread(target=fire) for _ in range(20)]
        for t in threads:
            t.start()
        for t in threads:
            t.join()

        accepted = sum(1 for r in results if r.accepted)
        self.assertEqual(accepted, MAX_ROLLBACKS_PER_WINDOW)

    def test_accepted_rollbacks_are_written_to_the_audit_log(self):
        base = 1_000_000.0
        for i in range(MAX_ROLLBACKS_PER_WINDOW):
            self.limiter.try_rollback("svc-audit", now=base + i)
        self.limiter.try_rollback("svc-audit", now=base + MAX_ROLLBACKS_PER_WINDOW)  # rejected

        entries = [e for e in self.log.all_entries() if e["service_name"] == "svc-audit"]
        self.assertEqual(len(entries), MAX_ROLLBACKS_PER_WINDOW,
                          "only ACCEPTED rollbacks should be audit-logged, not rejected ones")

    def test_retry_after_shrinks_as_time_passes(self):
        base = 1_000_000.0
        for i in range(MAX_ROLLBACKS_PER_WINDOW):
            self.limiter.try_rollback("svc-a", now=base + i)
        r1 = self.limiter.try_rollback("svc-a", now=base + MAX_ROLLBACKS_PER_WINDOW)
        r2 = self.limiter.try_rollback("svc-a", now=base + MAX_ROLLBACKS_PER_WINDOW + 30)
        self.assertGreater(r1.retry_after_seconds, r2.retry_after_seconds)


if __name__ == "__main__":
    unittest.main()
