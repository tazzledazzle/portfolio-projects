"""
Problem 5: Change-Failure-Rate (CFR) Calculator
==========================================
Given a stream of deploy events (service, timestamp, outcome), maintain a
ROLLING 7-DAY change-failure-rate per service, queryable on demand:

    CFR = (failed + rolled_back deploys in the last 7 days) / (total deploys in the last 7 days)

Must also handle the realistic case that a deploy's outcome isn't known
at insert time — it starts as PENDING and gets updated to SUCCESS /
FAILED / ROLLED_BACK later (e.g. a rollback triggered hours after a bad
deploy, once an alert fires).

Run:
    python3 -m unittest p5_change_failure_rate -v
"""
from __future__ import annotations

import itertools
import threading
from collections import defaultdict, deque
from dataclasses import dataclass
from enum import Enum
from typing import Optional


class Outcome(str, Enum):
    PENDING = "PENDING"
    SUCCESS = "SUCCESS"
    FAILED = "FAILED"
    ROLLED_BACK = "ROLLED_BACK"


def _is_failure(outcome: Outcome) -> bool:
    return outcome in (Outcome.FAILED, Outcome.ROLLED_BACK)


WINDOW_SECONDS = 7 * 24 * 60 * 60


@dataclass
class DeployEvent:
    event_id: int
    service_name: str
    timestamp: float
    outcome: Outcome


# ---------------------------------------------------------------------------
# STEP 1 — BRUTE FORCE
# ---------------------------------------------------------------------------
# Reasoning:
#   Store every event ever recorded in one flat list, and just append to
#   it (or, for the late-arriving-outcome case, mutate the matching
#   record in place by scanning for it). Every query re-scans the ENTIRE
#   list, filters to the requested service and the last 7 days, and
#   recomputes the ratio from scratch: O(n) per query, O(n) per outcome
#   update (to find the event to mutate), where n = ALL events ever
#   recorded, across ALL services, forever. This is the naive-but-obvious
#   version, and it's the one people ship first because it's simplest to
#   read and gets the ARITHMETIC right — the problem is purely that it
#   does not scale, and CFR is exactly the kind of metric that gets
#   queried very frequently (dashboards polling every few seconds, an
#   automated rollout gate checking it before every promotion).
# ---------------------------------------------------------------------------
class BruteForceCFRCalculator:
    def __init__(self):
        self._events: list[DeployEvent] = []
        self._next_id = itertools.count(1)

    def record_deploy(self, service_name: str, timestamp: float,
                       outcome: Outcome = Outcome.PENDING) -> int:
        event_id = next(self._next_id)
        self._events.append(DeployEvent(event_id, service_name, timestamp, outcome))
        return event_id

    def update_outcome(self, event_id: int, outcome: Outcome) -> None:
        # O(n) scan to find the event.
        for e in self._events:
            if e.event_id == event_id:
                e.outcome = outcome
                return
        raise KeyError(f"no such event_id: {event_id}")

    def change_failure_rate(self, service_name: str, now: float) -> Optional[float]:
        window_start = now - WINDOW_SECONDS
        # O(n) scan of ALL events across ALL services, every query.
        relevant = [
            e for e in self._events
            if e.service_name == service_name and window_start <= e.timestamp <= now
            and e.outcome != Outcome.PENDING  # a still-pending deploy has no verdict yet
        ]
        if not relevant:
            return None
        failures = sum(1 for e in relevant if _is_failure(e.outcome))
        return failures / len(relevant)


# ---------------------------------------------------------------------------
# STEP 2 — OPTIMAL
# ---------------------------------------------------------------------------
# Reasoning:
#   Two independent fixes, matching the two O(n) hot spots above:
#
#   (1) Per-service, not global. Partition storage by service_name up
#       front (a dict of deques), so a query for one service never has
#       to look at another service's events at all. With Brex's ~1,000
#       microservices, this alone turns an O(total events across the
#       company) scan into an O(this service's events) scan.
#
#   (2) O(1) amortized eviction + O(1) query via running counters. For
#       each service, maintain (a) a deque of that service's events in
#       time order, and (b) two running counters: total_resolved and
#       total_failures within the current window. On insert, we do NOT
#       add to the counters yet (the event starts PENDING, and a
#       PENDING event contributes to neither the numerator nor the
#       denominator — see below). On evict (an event aging out of the
#       7-day window), subtract its contribution if it had a resolved
#       outcome. change_failure_rate() becomes an O(1) counter read
#       after evicting expired events (amortized O(1) since each event
#       is evicted at most once, ever).
#
#   (3) The late-arriving-outcome problem needs an INDEX, not a scan:
#       update_outcome() must find "the event with this id" in O(1),
#       not O(n) — so we keep a separate event_id -> event object map.
#       When an outcome update lands, we adjust the running counters
#       IF AND ONLY IF that event is still inside the current 7-day
#       window (an outcome update arriving for an event that's already
#       aged out doesn't matter to a "last 7 days" metric, and skipping
#       that update is correct behavior, not a bug to special-case
#       around).
#
#   A subtlety worth naming explicitly for the interview: a PENDING
#   deploy is NOT counted in either the numerator or denominator until
#   it resolves. This is a deliberate modeling choice — counting it as
#   an implicit "success" until proven otherwise would make CFR look
#   artificially good right after a burst of deploys (before their
#   outcomes are known), which is exactly the wrong bias for a safety
#   metric. The tradeoff: CFR for "right now" is necessarily a lagging,
#   slightly-incomplete number, and that's the correct tradeoff for this
#   metric, not a bug to fix.
# ---------------------------------------------------------------------------
class OptimalCFRCalculator:
    def __init__(self):
        self._per_service: dict[str, deque[DeployEvent]] = defaultdict(deque)
        self._resolved_count: dict[str, int] = defaultdict(int)
        self._failure_count: dict[str, int] = defaultdict(int)
        self._by_id: dict[int, DeployEvent] = {}
        self._next_id = itertools.count(1)
        self._lock = threading.Lock()

    def record_deploy(self, service_name: str, timestamp: float,
                       outcome: Outcome = Outcome.PENDING) -> int:
        with self._lock:
            event_id = next(self._next_id)
            event = DeployEvent(event_id, service_name, timestamp, outcome)
            self._per_service[service_name].append(event)
            self._by_id[event_id] = event
            if outcome != Outcome.PENDING:
                self._resolved_count[service_name] += 1
                if _is_failure(outcome):
                    self._failure_count[service_name] += 1
            return event_id

    def update_outcome(self, event_id: int, outcome: Outcome) -> None:
        with self._lock:
            event = self._by_id.get(event_id)
            if event is None:
                raise KeyError(f"no such event_id: {event_id}")

            was_resolved = event.outcome != Outcome.PENDING
            was_failure = _is_failure(event.outcome)
            event.outcome = outcome
            is_resolved = outcome != Outcome.PENDING
            is_failure = _is_failure(outcome)

            svc = event.service_name
            if was_resolved and not is_resolved:
                self._resolved_count[svc] -= 1
            elif not was_resolved and is_resolved:
                self._resolved_count[svc] += 1
            if was_failure and not is_failure:
                self._failure_count[svc] -= 1
            elif not was_failure and is_failure:
                self._failure_count[svc] += 1

    def _evict_expired(self, service_name: str, now: float) -> None:
        window_start = now - WINDOW_SECONDS
        window = self._per_service[service_name]
        while window and window[0].timestamp < window_start:
            e = window.popleft()
            del self._by_id[e.event_id]
            if e.outcome != Outcome.PENDING:
                self._resolved_count[service_name] -= 1
                if _is_failure(e.outcome):
                    self._failure_count[service_name] -= 1

    def change_failure_rate(self, service_name: str, now: float) -> Optional[float]:
        with self._lock:
            self._evict_expired(service_name, now)
            resolved = self._resolved_count.get(service_name, 0)
            if resolved == 0:
                return None
            failures = self._failure_count.get(service_name, 0)
            return failures / resolved


# ---------------------------------------------------------------------------
# Tests
# ---------------------------------------------------------------------------
import unittest  # noqa: E402


class _CFRContractTests:
    IMPL = None

    def make(self):
        return self.IMPL()

    def test_no_events_returns_none(self):
        calc = self.make()
        self.assertIsNone(calc.change_failure_rate("svc", now=1000))

    def test_all_success_is_zero_percent(self):
        calc = self.make()
        for i in range(5):
            calc.record_deploy("svc", timestamp=1000 + i, outcome=Outcome.SUCCESS)
        self.assertEqual(calc.change_failure_rate("svc", now=1010), 0.0)

    def test_all_failed_is_one_hundred_percent(self):
        calc = self.make()
        for i in range(5):
            calc.record_deploy("svc", timestamp=1000 + i, outcome=Outcome.FAILED)
        self.assertEqual(calc.change_failure_rate("svc", now=1010), 1.0)

    def test_mixed_outcomes_rate(self):
        calc = self.make()
        calc.record_deploy("svc", 1000, Outcome.SUCCESS)
        calc.record_deploy("svc", 1001, Outcome.SUCCESS)
        calc.record_deploy("svc", 1002, Outcome.FAILED)
        calc.record_deploy("svc", 1003, Outcome.ROLLED_BACK)
        self.assertEqual(calc.change_failure_rate("svc", now=1010), 0.5)

    def test_pending_deploys_are_excluded_from_the_calculation(self):
        calc = self.make()
        calc.record_deploy("svc", 1000, Outcome.SUCCESS)
        calc.record_deploy("svc", 1001, Outcome.PENDING)
        calc.record_deploy("svc", 1002, Outcome.PENDING)
        # Only one resolved deploy, and it succeeded -> CFR 0.0, not
        # diluted or inflated by the two still-pending deploys.
        self.assertEqual(calc.change_failure_rate("svc", now=1010), 0.0)

    def test_events_outside_the_window_are_excluded(self):
        calc = self.make()
        calc.record_deploy("svc", timestamp=0, outcome=Outcome.FAILED)  # ancient, out of window
        calc.record_deploy("svc", timestamp=WINDOW_SECONDS + 100, outcome=Outcome.SUCCESS)
        rate = calc.change_failure_rate("svc", now=WINDOW_SECONDS + 100)
        self.assertEqual(rate, 0.0, "the old FAILED deploy must not count toward current CFR")

    def test_window_is_closed_interval_boundary_included_one_second_past_excluded(self):
        # Documented semantics: the window is [now - WINDOW_SECONDS, now],
        # closed on both ends — a deploy exactly WINDOW_SECONDS old still
        # counts (this test), but one second older than that does not
        # (the next test). Getting this boundary choice WRONG in either
        # direction is a classic off-by-one that silently changes what
        # "last 7 days" means by one data point right at the edge.
        calc = self.make()
        now = 1_000_000.0
        calc.record_deploy("svc", timestamp=now - WINDOW_SECONDS, outcome=Outcome.FAILED)
        rate = calc.change_failure_rate("svc", now=now)
        self.assertEqual(rate, 1.0, "an event exactly WINDOW_SECONDS old is still inside the window")

    def test_event_one_second_older_than_the_window_is_excluded(self):
        calc = self.make()
        now = 1_000_000.0
        calc.record_deploy("svc", timestamp=now - WINDOW_SECONDS - 1, outcome=Outcome.FAILED)
        calc.record_deploy("svc", timestamp=now - WINDOW_SECONDS + 1, outcome=Outcome.SUCCESS)
        rate = calc.change_failure_rate("svc", now=now)
        self.assertEqual(rate, 0.0, "an event older than WINDOW_SECONDS must not count")

    def test_services_are_independent(self):
        calc = self.make()
        calc.record_deploy("svc-a", 1000, Outcome.FAILED)
        calc.record_deploy("svc-b", 1000, Outcome.SUCCESS)
        self.assertEqual(calc.change_failure_rate("svc-a", now=1010), 1.0)
        self.assertEqual(calc.change_failure_rate("svc-b", now=1010), 0.0)

    def test_late_arriving_outcome_updates_the_rate(self):
        calc = self.make()
        eid = calc.record_deploy("svc", 1000, Outcome.PENDING)
        calc.record_deploy("svc", 1001, Outcome.SUCCESS)
        # Right now: only 1 resolved deploy (SUCCESS) -> CFR 0.0
        self.assertEqual(calc.change_failure_rate("svc", now=1010), 0.0)

        # Hours later, an alert fires and the pending deploy is
        # discovered to have actually failed and been rolled back.
        calc.update_outcome(eid, Outcome.ROLLED_BACK)
        self.assertEqual(calc.change_failure_rate("svc", now=1010), 0.5)

    def test_outcome_can_be_corrected_twice(self):
        calc = self.make()
        eid = calc.record_deploy("svc", 1000, Outcome.PENDING)
        calc.update_outcome(eid, Outcome.FAILED)
        self.assertEqual(calc.change_failure_rate("svc", now=1010), 1.0)
        # A human manually re-triages it as actually fine (false alarm).
        calc.update_outcome(eid, Outcome.SUCCESS)
        self.assertEqual(calc.change_failure_rate("svc", now=1010), 0.0)

    def test_update_unknown_event_id_raises(self):
        calc = self.make()
        with self.assertRaises(KeyError):
            calc.update_outcome(99999, Outcome.FAILED)

    def test_rolling_window_forgets_old_deploys_over_time(self):
        calc = self.make()
        calc.record_deploy("svc", timestamp=0, outcome=Outcome.FAILED)
        self.assertEqual(calc.change_failure_rate("svc", now=1000), 1.0)
        # Seven days and one second later, that failed deploy should no
        # longer count at all.
        far_future = WINDOW_SECONDS + 1
        calc.record_deploy("svc", timestamp=far_future, outcome=Outcome.SUCCESS)
        self.assertEqual(calc.change_failure_rate("svc", now=far_future), 0.0)


class TestBruteForceCFR(_CFRContractTests, unittest.TestCase):
    IMPL = BruteForceCFRCalculator


class TestOptimalCFR(_CFRContractTests, unittest.TestCase):
    IMPL = OptimalCFRCalculator

    def test_matches_brute_force_across_a_long_randomized_stream(self):
        import random
        random.seed(7)
        opt = self.make()
        brute = BruteForceCFRCalculator()
        services = ["svc-a", "svc-b", "svc-c"]
        outcomes = [Outcome.SUCCESS, Outcome.FAILED, Outcome.ROLLED_BACK, Outcome.PENDING]

        now = 0.0
        recorded_ids: list[tuple[int, int]] = []  # (opt_id, brute_id) pairs, same index
        for _ in range(300):
            now += random.uniform(1, WINDOW_SECONDS / 20)
            svc = random.choice(services)
            outcome = random.choice(outcomes)
            oid = opt.record_deploy(svc, now, outcome)
            bid = brute.record_deploy(svc, now, outcome)
            recorded_ids.append((oid, bid))

            # Occasionally resolve a previously-pending event late.
            if recorded_ids and random.random() < 0.2:
                idx = random.randrange(len(recorded_ids))
                oid2, bid2 = recorded_ids[idx]
                new_outcome = random.choice([Outcome.SUCCESS, Outcome.FAILED, Outcome.ROLLED_BACK])
                try:
                    opt.update_outcome(oid2, new_outcome)
                    brute.update_outcome(bid2, new_outcome)
                except KeyError:
                    pass  # event already evicted from one side's window bookkeeping is fine

            for svc in services:
                self.assertEqual(
                    opt.change_failure_rate(svc, now),
                    brute.change_failure_rate(svc, now),
                    f"mismatch for {svc} at t={now}",
                )

    def test_concurrent_inserts_do_not_corrupt_counters(self):
        calc = self.make()
        barrier = threading.Barrier(20)

        def fire(i):
            barrier.wait()
            calc.record_deploy("svc-concurrent", timestamp=1000.0 + i * 0.001,
                                outcome=Outcome.FAILED if i % 2 == 0 else Outcome.SUCCESS)

        threads = [threading.Thread(target=fire, args=(i,)) for i in range(20)]
        for t in threads:
            t.start()
        for t in threads:
            t.join()

        rate = calc.change_failure_rate("svc-concurrent", now=2000)
        self.assertEqual(rate, 0.5)


if __name__ == "__main__":
    unittest.main()
