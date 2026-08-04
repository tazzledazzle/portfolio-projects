"""
Problem 3: Canary Analysis Evaluator
==========================================
Given a stream of (timestamp, error_count, request_count) tuples for a
canary deployment and for the baseline, decide PASS / FAIL / CONTINUE
based on whether the canary's error rate exceeds the baseline's error
rate by more than a threshold, over a rolling time window. Must handle
sparse/missing data points gracefully.

Run:
    python3 -m unittest p3_canary_evaluator -v
"""
from __future__ import annotations

from collections import deque
from dataclasses import dataclass
from enum import Enum
from typing import Optional


class Verdict(str, Enum):
    PASS = "PASS"
    FAIL = "FAIL"
    CONTINUE = "CONTINUE"  # not enough data yet to decide either way


@dataclass(frozen=True)
class MetricPoint:
    timestamp: float
    error_count: int
    request_count: int


# Design decisions shared by both implementations:
#  - error rate is undefined (0/0) when request_count == 0; we treat a
#    window with zero total requests as "insufficient data" -> CONTINUE,
#    never as 0% error (that would be a false PASS on a canary nobody
#    is even sending traffic to).
#  - "sparse data" is handled by defining the window in WALL-CLOCK TIME,
#    not in "last K points" — a canary that's only getting one request
#    every 90 seconds still gets evaluated correctly over, say, a 5-minute
#    window, instead of a count-based window silently spanning hours.
#  - MIN_SAMPLE_SIZE guards against a statistically meaningless verdict:
#    2 requests with 1 error is a 50% error rate by arithmetic, but it's
#    noise, not signal. Below the threshold -> CONTINUE.
WINDOW_SECONDS = 5 * 60
MIN_SAMPLE_SIZE = 50          # per side (canary/baseline), within the window
FAIL_THRESHOLD_DELTA = 0.05   # canary error rate must not exceed baseline by more than 5 pts


def _error_rate(error_count: int, request_count: int) -> Optional[float]:
    if request_count == 0:
        return None
    return error_count / request_count


def _decide(canary_errors: int, canary_requests: int,
            baseline_errors: int, baseline_requests: int) -> Verdict:
    if canary_requests < MIN_SAMPLE_SIZE or baseline_requests < MIN_SAMPLE_SIZE:
        return Verdict.CONTINUE

    canary_rate = _error_rate(canary_errors, canary_requests)
    baseline_rate = _error_rate(baseline_errors, baseline_requests)
    if canary_rate is None or baseline_rate is None:
        return Verdict.CONTINUE

    delta = canary_rate - baseline_rate
    if delta > FAIL_THRESHOLD_DELTA:
        return Verdict.FAIL
    return Verdict.PASS


# ---------------------------------------------------------------------------
# STEP 1 — BRUTE FORCE
# ---------------------------------------------------------------------------
# Reasoning:
#   Keep every point ever received in two plain lists (canary, baseline).
#   On every `evaluate(now)` call, filter BOTH lists down to points inside
#   [now - WINDOW_SECONDS, now], sum errors and requests across whatever's
#   left, and run the decision function. Correct, and easy to reason
#   about — but it's O(n) per evaluation where n = total points ever
#   received, and that cost only grows: a canary that's been running for
#   an hour with a data point every second means every single evaluation
#   after minute one is re-scanning thousands of points to reconstruct a
#   5-minute window. For an evaluator that's likely being polled every
#   few seconds by an automated rollout controller, that's wasteful and
#   eventually becomes the bottleneck itself.
# ---------------------------------------------------------------------------
class BruteForceCanaryEvaluator:
    def __init__(self):
        self._canary_points: list[MetricPoint] = []
        self._baseline_points: list[MetricPoint] = []

    def record_canary(self, timestamp: float, error_count: int, request_count: int) -> None:
        self._canary_points.append(MetricPoint(timestamp, error_count, request_count))

    def record_baseline(self, timestamp: float, error_count: int, request_count: int) -> None:
        self._baseline_points.append(MetricPoint(timestamp, error_count, request_count))

    def evaluate(self, now: float) -> Verdict:
        window_start = now - WINDOW_SECONDS

        # O(n) scan + rebuild, every call.
        canary_in_window = [p for p in self._canary_points if window_start <= p.timestamp <= now]
        baseline_in_window = [p for p in self._baseline_points if window_start <= p.timestamp <= now]

        c_errors = sum(p.error_count for p in canary_in_window)
        c_requests = sum(p.request_count for p in canary_in_window)
        b_errors = sum(p.error_count for p in baseline_in_window)
        b_requests = sum(p.request_count for p in baseline_in_window)

        return _decide(c_errors, c_requests, b_errors, b_requests)


# ---------------------------------------------------------------------------
# STEP 2 — OPTIMAL
# ---------------------------------------------------------------------------
# Reasoning:
#   Maintain a deque PER SIDE (canary/baseline) plus running sums of
#   error_count and request_count for exactly what's currently "in the
#   window". On each new point: add it to the running sums and push it
#   onto the right of the deque. Before evaluating (or whenever a new
#   point arrives), evict expired points from the LEFT of the deque and
#   subtract their contribution from the running sums. Since every point
#   is pushed once and popped at most once over its lifetime, this is
#   amortized O(1) per recorded point and O(1) per evaluation — instead
#   of O(n) per evaluation.
#
#   This assumes points arrive in roughly non-decreasing timestamp order
#   per side (true for a live metrics stream feeding a canary evaluator).
#   If out-of-order points needed to be supported, we'd need a different
#   structure (e.g., a time-bucketed histogram) — noted in the test
#   matrix as a case that's explicitly OUT of scope for this design, with
#   the tradeoff called out rather than silently ignored.
# ---------------------------------------------------------------------------
class OptimalCanaryEvaluator:
    def __init__(self):
        self._canary_window: deque[MetricPoint] = deque()
        self._baseline_window: deque[MetricPoint] = deque()
        self._c_errors = 0
        self._c_requests = 0
        self._b_errors = 0
        self._b_requests = 0

    def record_canary(self, timestamp: float, error_count: int, request_count: int) -> None:
        self._canary_window.append(MetricPoint(timestamp, error_count, request_count))
        self._c_errors += error_count
        self._c_requests += request_count

    def record_baseline(self, timestamp: float, error_count: int, request_count: int) -> None:
        self._baseline_window.append(MetricPoint(timestamp, error_count, request_count))
        self._b_errors += error_count
        self._b_requests += request_count

    def _evict_expired(self, now: float) -> None:
        window_start = now - WINDOW_SECONDS
        while self._canary_window and self._canary_window[0].timestamp < window_start:
            p = self._canary_window.popleft()
            self._c_errors -= p.error_count
            self._c_requests -= p.request_count
        while self._baseline_window and self._baseline_window[0].timestamp < window_start:
            p = self._baseline_window.popleft()
            self._b_errors -= p.error_count
            self._b_requests -= p.request_count

    def evaluate(self, now: float) -> Verdict:
        self._evict_expired(now)
        return _decide(self._c_errors, self._c_requests, self._b_errors, self._b_requests)


# ---------------------------------------------------------------------------
# Tests
# ---------------------------------------------------------------------------
import unittest  # noqa: E402


class _EvaluatorContractTests:
    """Shared test contract — run against BOTH implementations to prove
    they agree on every case, not just that the optimal one 'looks right'."""
    IMPL = None  # set by subclasses

    def make(self):
        return self.IMPL()

    def test_insufficient_samples_is_continue_not_pass_or_fail(self):
        ev = self.make()
        ev.record_canary(0, error_count=1, request_count=2)   # way below MIN_SAMPLE_SIZE
        ev.record_baseline(0, error_count=0, request_count=2)
        self.assertEqual(ev.evaluate(now=0), Verdict.CONTINUE)

    def test_zero_traffic_on_either_side_is_continue(self):
        ev = self.make()
        ev.record_canary(0, error_count=0, request_count=0)
        ev.record_baseline(0, error_count=0, request_count=100)
        self.assertEqual(ev.evaluate(now=0), Verdict.CONTINUE,
                          "a canary with zero traffic must not be treated as a 0% error rate PASS")

    def test_equal_error_rates_pass(self):
        ev = self.make()
        ev.record_canary(0, error_count=5, request_count=100)
        ev.record_baseline(0, error_count=5, request_count=100)
        self.assertEqual(ev.evaluate(now=0), Verdict.PASS)

    def test_canary_worse_by_exactly_threshold_is_pass_boundary(self):
        # delta must be STRICTLY greater than threshold to fail (not >=),
        # so exactly-at-threshold is a PASS. This is a deliberate,
        # documented boundary choice, not an accident.
        ev = self.make()
        ev.record_canary(0, error_count=10, request_count=100)   # 10%
        ev.record_baseline(0, error_count=5, request_count=100)  # 5% -> delta exactly 0.05
        self.assertEqual(ev.evaluate(now=0), Verdict.PASS)

    def test_canary_worse_by_threshold_plus_epsilon_fails(self):
        ev = self.make()
        ev.record_canary(0, error_count=11, request_count=100)   # 11%
        ev.record_baseline(0, error_count=5, request_count=100)  # 5% -> delta 0.06 > 0.05
        self.assertEqual(ev.evaluate(now=0), Verdict.FAIL)

    def test_canary_better_than_baseline_passes(self):
        ev = self.make()
        ev.record_canary(0, error_count=1, request_count=100)
        ev.record_baseline(0, error_count=10, request_count=100)
        self.assertEqual(ev.evaluate(now=0), Verdict.PASS)

    def test_points_outside_window_are_ignored(self):
        ev = self.make()
        # A huge spike LONG before the window: must not affect the verdict.
        ev.record_canary(timestamp=0, error_count=1000, request_count=1000)
        # Healthy, current traffic, inside the window.
        for t in range(1000, 1000 + 60):
            ev.record_canary(timestamp=t, error_count=0, request_count=2)
            ev.record_baseline(timestamp=t, error_count=0, request_count=2)
        verdict = ev.evaluate(now=1000 + 60)
        self.assertEqual(verdict, Verdict.PASS)

    def test_sparse_data_still_evaluates_correctly(self):
        # One data point every 90s instead of every second — still must
        # accumulate correctly across the 5-minute window.
        ev = self.make()
        for i, t in enumerate([0, 90, 180, 270]):
            ev.record_canary(t, error_count=1, request_count=20)
            ev.record_baseline(t, error_count=1, request_count=20)
        verdict = ev.evaluate(now=270)
        # 4 * 20 = 80 requests each side, still under MIN_SAMPLE_SIZE... bump it up:
        for i, t in enumerate([300, 330]):
            ev.record_canary(t, error_count=1, request_count=20)
            ev.record_baseline(t, error_count=1, request_count=20)
        verdict = ev.evaluate(now=330)
        self.assertIn(verdict, (Verdict.PASS, Verdict.FAIL),
                       "with enough sparse points accumulated, a real verdict should be reachable")


class TestBruteForceEvaluator(_EvaluatorContractTests, unittest.TestCase):
    IMPL = BruteForceCanaryEvaluator


class TestOptimalEvaluator(_EvaluatorContractTests, unittest.TestCase):
    IMPL = OptimalCanaryEvaluator

    def test_optimal_running_sums_match_a_manual_recount(self):
        """Directly checks that the incremental running sums the optimal
        version maintains equal a brute-force recount at every step —
        i.e. proves the O(1) incremental approach and the O(n) approach
        never disagree, across a long stream with points aging out."""
        ev = self.make()
        brute = BruteForceCanaryEvaluator()

        import random
        random.seed(42)
        now = 0.0
        for _ in range(500):
            now += random.uniform(0.1, 3.0)
            err, req = random.randint(0, 3), random.randint(1, 20)
            ev.record_canary(now, err, req)
            brute.record_canary(now, err, req)
            err2, req2 = random.randint(0, 3), random.randint(1, 20)
            ev.record_baseline(now, err2, req2)
            brute.record_baseline(now, err2, req2)

            self.assertEqual(ev.evaluate(now), brute.evaluate(now),
                              f"mismatch at t={now}")


if __name__ == "__main__":
    unittest.main()
