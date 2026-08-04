"""
Category C -- weighted interval scheduling / max-profit pattern.
Run: python3 variations/kata_c_scheduling.py
"""

from __future__ import annotations

import bisect
from dataclasses import dataclass, field


# C1 ---------------------------------------------------------------------
def max_value_spot_instance_reservations(bids: list[dict]) -> int:
    """bids: [{"start": int, "end": int, "bid_value": int}] -> max total value."""
    bids = sorted(bids, key=lambda b: b["end"])
    ends = [b["end"] for b in bids]
    n = len(bids)
    dp = [0] * (n + 1)
    for i in range(1, n + 1):
        b = bids[i - 1]
        p = bisect.bisect_right(ends, b["start"], 0, i - 1)
        dp[i] = max(dp[i - 1], dp[p] + b["bid_value"])
    return dp[n]


# C2 ---------------------------------------------------------------------
def max_revenue_ad_slot_booking(campaigns: list[dict]) -> tuple[int, list[dict], list[dict]]:
    """Returns (max_revenue, accepted, rejected)."""
    ordered = sorted(campaigns, key=lambda c: c["window_end"])
    ends = [c["window_end"] for c in ordered]
    n = len(ordered)
    dp = [0] * (n + 1)
    choice = [False] * (n + 1)
    for i in range(1, n + 1):
        c = ordered[i - 1]
        p = bisect.bisect_right(ends, c["window_start"], 0, i - 1)
        take = dp[p] + c["revenue"]
        if take > dp[i - 1]:
            dp[i], choice[i] = take, True
        else:
            dp[i] = dp[i - 1]
    accepted = []
    i = n
    while i > 0:
        if choice[i]:
            c = ordered[i - 1]
            accepted.append(c)
            i = bisect.bisect_right(ends, c["window_start"], 0, i - 1)
        else:
            i -= 1
    accepted.reverse()
    accepted_ids = {id(c) for c in accepted}
    rejected = [c for c in ordered if id(c) not in accepted_ids]
    return dp[n], accepted, rejected


# C3 ---------------------------------------------------------------------
def max_priority_room_bookings(bookings: list[dict]) -> int:
    """
    Same as base pattern, but end==start counts as overlapping (no
    back-to-back). Uses bisect_left instead of bisect_right against `start`
    so an exact end==start match is excluded from the eligible predecessors.
    """
    ordered = sorted(bookings, key=lambda b: b["end"])
    ends = [b["end"] for b in ordered]
    n = len(ordered)
    dp = [0] * (n + 1)
    for i in range(1, n + 1):
        b = ordered[i - 1]
        # strictly earlier end required now: use bisect_left
        p = bisect.bisect_left(ends, b["start"], 0, i - 1)
        dp[i] = max(dp[i - 1], dp[p] + b["priority"])
    return dp[n]


# C4 ---------------------------------------------------------------------
class GigScheduler:
    """
    Precompute the DP once; answer 'max payout using gigs ending by T'
    queries in O(log n) each.
    """

    def __init__(self, gigs: list[dict]):
        self.gigs = sorted(gigs, key=lambda g: g["end"])
        self.ends = [g["end"] for g in self.gigs]
        n = len(self.gigs)
        self.dp = [0] * (n + 1)
        for i in range(1, n + 1):
            g = self.gigs[i - 1]
            p = bisect.bisect_right(self.ends, g["start"], 0, i - 1)
            self.dp[i] = max(self.dp[i - 1], self.dp[p] + g["payout"])

    def max_payout(self) -> int:
        return self.dp[-1]

    def max_payout_ending_by(self, t: int) -> int:
        idx = bisect.bisect_right(self.ends, t)
        return self.dp[idx]


# C5 ---------------------------------------------------------------------
def max_stipend_oncall_shift_bidding(bids: list[dict], max_shifts_per_engineer: int = 2) -> int:
    """
    Adds a per-engineer cap, which breaks the clean 1-D DP: the optimal
    subproblem now needs (index, remaining_capacity_per_engineer), which is
    combinatorial in the number of engineers if done exactly.

    Practical approach implemented here: DP over (index, shift-count-so-far
    for the *specific engineer of job i*) is still not enough because the
    state needs per-engineer counts, not a single count. We solve this with
    a bounded state: dp[i][c] where c = how many of THIS shift's engineer's
    slots remain, recomputed via a helper that tracks per-engineer usage
    with a dict during a single top-down pass with memoization keyed on
    (index, frozenset of engineer counts) — exponential in the worst case,
    but fine for interview-scale n (~dozens) and is presented here as the
    "correct but not production-scale" solution, with the trade-off called
    out explicitly rather than hidden.
    """
    from functools import lru_cache

    ordered = sorted(bids, key=lambda b: b["end"])
    ends = [b["end"] for b in ordered]
    n = len(ordered)
    pred = [bisect.bisect_right(ends, ordered[i]["start"], 0, i) for i in range(n)]

    def counts_key(counts: dict[str, int]) -> tuple:
        return tuple(sorted(counts.items()))

    @lru_cache(maxsize=None)
    def solve(i: int, counts_frozen: tuple) -> int:
        if i < 0:
            return 0
        counts = dict(counts_frozen)
        skip = solve(i - 1, counts_frozen)
        eng = ordered[i]["engineer_id"]
        if counts.get(eng, 0) < max_shifts_per_engineer:
            new_counts = dict(counts)
            new_counts[eng] = new_counts.get(eng, 0) + 1
            p = pred[i] - 1  # convert to 0-indexed "last eligible job index"
            take = ordered[i]["stipend"] + solve(p, counts_key(new_counts))
        else:
            take = float("-inf")
        return max(skip, take)

    return solve(n - 1, ())


# C6 ---------------------------------------------------------------------
def max_profit_with_setup_time(jobs: list[dict], setup_minutes: int) -> int:
    """job[i] eligible after job[j] only if job[i].start >= job[j].end + setup_minutes."""
    ordered = sorted(jobs, key=lambda j: j["end"])
    ends = [j["end"] for j in ordered]
    n = len(ordered)
    dp = [0] * (n + 1)
    for i in range(1, n + 1):
        j = ordered[i - 1]
        # only change from the base problem: subtract setup_minutes from the bound
        p = bisect.bisect_right(ends, j["start"] - setup_minutes, 0, i - 1)
        dp[i] = max(dp[i - 1], dp[p] + j["profit"])
    return dp[n]


def _run_tests():
    bids = [
        {"start": 1, "end": 4, "bid_value": 3},
        {"start": 3, "end": 5, "bid_value": 5},
        {"start": 0, "end": 6, "bid_value": 10},
    ]
    assert max_value_spot_instance_reservations(bids) == 10
    print("C1 max_value_spot_instance_reservations: OK")

    campaigns = [
        {"window_start": 1, "window_end": 3, "revenue": 5},
        {"window_start": 2, "window_end": 5, "revenue": 6},
        {"window_start": 4, "window_end": 6, "revenue": 5},
    ]
    total, accepted, rejected = max_revenue_ad_slot_booking(campaigns)
    assert total == 10  # (1,3,5)+(4,6,5)
    assert len(accepted) + len(rejected) == len(campaigns)
    print("C2 max_revenue_ad_slot_booking: OK")

    # touching bookings (end == start) must NOT both be accepted
    bookings = [
        {"start": 1, "end": 3, "priority": 5},
        {"start": 3, "end": 5, "priority": 5},  # touches previous
    ]
    assert max_priority_room_bookings(bookings) == 5
    print("C3 max_priority_room_bookings: OK")

    gigs = [
        {"start": 1, "end": 3, "payout": 5},
        {"start": 4, "end": 6, "payout": 5},
        {"start": 7, "end": 9, "payout": 5},
    ]
    scheduler = GigScheduler(gigs)
    assert scheduler.max_payout() == 15
    assert scheduler.max_payout_ending_by(6) == 10
    assert scheduler.max_payout_ending_by(2) == 0
    print("C4 GigScheduler: OK")

    shift_bids = [
        {"start": 0, "end": 8, "stipend": 100, "engineer_id": "eng1"},
        {"start": 8, "end": 16, "stipend": 100, "engineer_id": "eng1"},
        {"start": 16, "end": 24, "stipend": 100, "engineer_id": "eng1"},  # 3rd for eng1, capped at 2
        {"start": 16, "end": 24, "stipend": 90, "engineer_id": "eng2"},
    ]
    result = max_stipend_oncall_shift_bidding(shift_bids, max_shifts_per_engineer=2)
    assert result == 200 + 90, result  # eng1 takes shifts 1+2 (200), eng2 takes the 24h slot (90)
    print("C5 max_stipend_oncall_shift_bidding: OK")

    jobs = [
        {"start": 0, "end": 2, "profit": 5},
        {"start": 2, "end": 4, "profit": 5},   # no gap -- excluded if setup_minutes > 0
        {"start": 5, "end": 7, "profit": 5},   # 1-unit gap after (0,2)? no, after (2,4) gap=1
    ]
    no_setup = max_profit_with_setup_time(jobs, setup_minutes=0)
    with_setup = max_profit_with_setup_time(jobs, setup_minutes=1)
    assert no_setup >= with_setup
    print(f"C6 max_profit_with_setup_time: no_setup={no_setup}, with_setup={with_setup} OK")

    print("\nAll category C tests passed.")


if __name__ == "__main__":
    _run_tests()
