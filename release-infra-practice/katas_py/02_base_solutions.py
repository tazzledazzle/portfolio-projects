"""
Reference implementations for the 3 base problems.
Run directly: python3 02_base_solutions.py
Each section is self-contained with its own assertions.
"""

from __future__ import annotations

import bisect
from dataclasses import dataclass
from datetime import datetime, timezone
from typing import Optional


# ---------------------------------------------------------------------------
# 1. first_duplicate_deployment
# ---------------------------------------------------------------------------

def first_duplicate_deployment(deployments: list[dict]) -> Optional[str]:
    """
    deployments: list of {"id": str, "timestamp": str (ISO 8601), ...}
    Returns the id of the first deployment record whose id repeats an
    earlier one in stream order, or None if all ids are unique.

    Time:  O(n) single pass, O(1) avg per hash-set op
    Space: O(n) worst case for the `seen` set
    """
    seen: set[str] = set()
    for record in deployments:
        dep_id = record["id"]
        if dep_id in seen:
            return dep_id
        seen.add(dep_id)
    return None


def _test_first_duplicate_deployment():
    deployments = [
        {"id": "d1", "timestamp": "2026-01-01T00:00:00Z"},
        {"id": "d2", "timestamp": "2026-01-01T00:05:00Z"},
        {"id": "d3", "timestamp": "2026-01-01T00:10:00Z"},
        {"id": "d2", "timestamp": "2026-01-01T00:15:00Z"},  # first repeat
        {"id": "d1", "timestamp": "2026-01-01T00:20:00Z"},
    ]
    assert first_duplicate_deployment(deployments) == "d2"

    all_unique = [{"id": "d1"}, {"id": "d2"}, {"id": "d3"}]
    assert first_duplicate_deployment(all_unique) is None

    empty: list[dict] = []
    assert first_duplicate_deployment(empty) is None
    print("first_duplicate_deployment: OK")


# ---------------------------------------------------------------------------
# 2. parse_and_sort_events
# ---------------------------------------------------------------------------

@dataclass
class Event:
    name: str
    timestamp: datetime


def _parse_timestamp(raw: str) -> datetime:
    """
    Normalize a handful of realistic API timestamp formats to a UTC-aware
    datetime. This is the part of the problem worth the most care.
    """
    raw = raw.strip()
    # Handle trailing 'Z' explicitly for portability across Python versions.
    if raw.endswith("Z"):
        raw = raw[:-1] + "+00:00"
    dt = datetime.fromisoformat(raw)
    if dt.tzinfo is None:
        dt = dt.replace(tzinfo=timezone.utc)
    return dt.astimezone(timezone.utc)


def parse_and_sort_events(raw_events: list[dict]) -> list[Event]:
    """
    raw_events: list of {"name": str, "timestamp": str (ISO 8601-ish)}
    Returns events parsed into typed form and sorted ascending by time.

    Time:  O(n log n) — sort dominates parsing
    Space: O(n) for the parsed list
    """
    parsed = [
        Event(name=r["name"], timestamp=_parse_timestamp(r["timestamp"]))
        for r in raw_events
    ]
    parsed.sort(key=lambda e: e.timestamp)
    return parsed


def _test_parse_and_sort_events():
    raw = [
        {"name": "deploy_started", "timestamp": "2026-01-01T00:10:00Z"},
        {"name": "build_queued", "timestamp": "2026-01-01T00:00:00+00:00"},
        {"name": "deploy_finished", "timestamp": "2026-01-01T00:20:00Z"},
    ]
    result = parse_and_sort_events(raw)
    assert [e.name for e in result] == [
        "build_queued",
        "deploy_started",
        "deploy_finished",
    ]
    assert result[0].timestamp < result[1].timestamp < result[2].timestamp
    print("parse_and_sort_events: OK")


# ---------------------------------------------------------------------------
# 3. max_profit_schedules
# ---------------------------------------------------------------------------

@dataclass
class Job:
    start: int
    end: int
    profit: int


def max_profit_schedules(jobs: list[Job]) -> tuple[int, list[Job]]:
    """
    Weighted interval scheduling: choose a subset of non-overlapping jobs
    that maximizes total profit. Returns (max_profit, chosen_jobs).

    Time:  O(n log n) — sort + binary search per job
    Space: O(n) for the DP table and choice-reconstruction array
    """
    jobs = sorted(jobs, key=lambda j: j.end)
    n = len(jobs)
    ends = [j.end for j in jobs]

    dp = [0] * (n + 1)
    choice = [False] * (n + 1)  # did we take job i-1 in the optimal solution?

    for i in range(1, n + 1):
        job = jobs[i - 1]
        # latest index p such that jobs[p-1].end <= job.start (1-indexed into dp)
        p = bisect.bisect_right(ends, job.start, 0, i - 1)
        take = dp[p] + job.profit
        skip = dp[i - 1]
        if take > skip:
            dp[i] = take
            choice[i] = True
        else:
            dp[i] = skip

    # reconstruct
    chosen: list[Job] = []
    i = n
    while i > 0:
        if choice[i]:
            job = jobs[i - 1]
            chosen.append(job)
            p = bisect.bisect_right(ends, job.start, 0, i - 1)
            i = p
        else:
            i -= 1
    chosen.reverse()
    return dp[n], chosen


def _test_max_profit_schedules():
    jobs = [
        Job(start=1, end=3, profit=5),
        Job(start=2, end=5, profit=6),
        Job(start=4, end=6, profit=5),
        Job(start=6, end=7, profit=4),
        Job(start=5, end=8, profit=11),
        Job(start=7, end=9, profit=2),
    ]
    profit, chosen = max_profit_schedules(jobs)
    assert profit == 17, profit
    print(f"max_profit_schedules: profit={profit}, chosen={chosen}")
    total = sum(j.profit for j in chosen)
    assert total == profit
    # verify no overlaps in chosen set
    for a, b in zip(chosen, chosen[1:]):
        assert a.end <= b.start
    print("max_profit_schedules: OK")


if __name__ == "__main__":
    _test_first_duplicate_deployment()
    _test_parse_and_sort_events()
    _test_max_profit_schedules()
    print("\nAll base problem tests passed.")
