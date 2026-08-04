"""
LeetCode 3389 - Minimum Operations to Make Character Frequencies Equal
========================================================================

Problem
-------
Given a string `s` of lowercase letters, you may repeatedly:
  1. Delete a character                              (cost 1)
  2. Insert a character                               (cost 1)
  3. Change a character to its *next* letter only      (cost 1)
     (no wraparound: 'z' cannot become 'a')

Return the minimum number of operations so that every distinct character
remaining in the string occurs the same number of times ("good" string;
the empty string counts as good).

Core idea
---------
Only the 26 letter *counts* matter (position is irrelevant). Pick a target
frequency `t` and an "active set" A of letters that will end up with
exactly `t` copies; every letter not in A must end up with 0 copies.

For letter i, let diff[i] = cnt[i] - target[i], where target[i] = t if
i in A else 0.
  - diff[i] > 0  -> "excess"  : that many copies must be removed
  - diff[i] < 0  -> "deficit" : that many copies must be added

Naively, cost = sum(|diff[i]|)  (delete every excess unit, insert every
deficit unit).

BUT operation 3 lets us convert one excess unit of letter i directly into
one deficit unit of letter i+1 for cost 1, instead of delete+insert (cost
2). This only works for *immediate* neighbors, and only forward
(i -> i+1), since letters can only be changed to their next letter.

So:
    cost(t, A) = sum(|diff[i]|) - sum_{i=0}^{24} min(excess[i], deficit[i+1])

Because a position's diff value has a single fixed sign, it can only ever
act as an excess *source* for i+1's deficit, or as a deficit *sink* for
i-1's excess -- never both -- so the pairwise mins don't interfere with
each other and can be summed independently, no complex matching needed.

We now need to search over t and over the active set A to minimize cost.

Solution 1: Brute force
------------------------
Directly enumerates every subset of "candidate" letters (letters that are
present in s, plus their immediate left/right neighbors -- letters with
count 0 that aren't adjacent to a present letter can *never* help, since
they'd only add pure insert cost with no possible discount) for every
target t. Exponential in the number of distinct letters, but simple and
obviously correct -- used here mainly to validate the optimized solution.

Solution 2: Optimal DP
------------------------
For a fixed t, instead of brute-forcing subsets, do a left-to-right DP
over the 26 letters. The only interaction between letters is the adjacent
pairwise saving, so the DP state is just "was the previous letter active
or not" (which fixes its excess value). This gives O(26) work per target
t, and t ranges over 1..max(cnt), so total time is O(26 * max_cnt).

    Time:  O(26 * max_cnt)   (max_cnt <= len(s) <= 2*10^4)
    Space: O(1)
"""

from __future__ import annotations

from collections import Counter
from itertools import combinations
from typing import Dict, List, Tuple

ALPHABET_SIZE = 26


def _counts(s: str) -> List[int]:
    """Return the 26-length frequency table for a lowercase string."""
    cnt = [0] * ALPHABET_SIZE
    for ch in s:
        cnt[ord(ch) - ord("a")] += 1
    return cnt


def _cost_for_active_set(cnt: List[int], t: int, active: set) -> int:
    """
    Given fixed counts, target frequency t, and a chosen active set of
    letter indices, compute the total operation cost using:
        cost = sum(|diff|) - sum_i min(excess[i], deficit[i+1])
    """
    diff = [cnt[i] - (t if i in active else 0) for i in range(ALPHABET_SIZE)]
    cost = sum(abs(d) for d in diff)
    for i in range(ALPHABET_SIZE - 1):
        excess_i = diff[i] if diff[i] > 0 else 0
        deficit_next = -diff[i + 1] if diff[i + 1] < 0 else 0
        cost -= min(excess_i, deficit_next)
    return cost


# ---------------------------------------------------------------------------
# Solution 1: Brute force (subset enumeration over candidate letters)
# ---------------------------------------------------------------------------
def min_operations_bruteforce(s: str) -> int:
    """
    Brute-force reference solution.

    For every target frequency t in [1, max_cnt], enumerate every subset
    of "candidate" letters (present letters + their immediate neighbors)
    and evaluate the exact cost formula. Track the global minimum. The
    all-letters-excluded subset (cost = len(s), i.e. delete everything)
    is automatically included as the empty subset, so t=0 / empty string
    is implicitly covered.

    Complexity: O(max_cnt * 2^k) where k = number of candidate letters
    (k <= 3 * distinct_letters, capped at 26). Only practical for small
    / low-diversity inputs -- that's expected for a brute force.
    """
    n = len(s)
    if n == 0:
        return 0

    cnt = _counts(s)
    present = [i for i in range(ALPHABET_SIZE) if cnt[i] > 0]

    candidates = set()
    for i in present:
        candidates.add(i)
        if i - 1 >= 0:
            candidates.add(i - 1)
        if i + 1 < ALPHABET_SIZE:
            candidates.add(i + 1)
    candidates = sorted(candidates)
    k = len(candidates)

    max_cnt = max(cnt)
    best = n  # baseline: delete every character

    for t in range(1, max_cnt + 1):
        for mask in range(1 << k):
            active = {candidates[j] for j in range(k) if mask & (1 << j)}
            cost = _cost_for_active_set(cnt, t, active)
            if cost < best:
                best = cost

    return best


# ---------------------------------------------------------------------------
# Solution 2: Optimal O(26 * max_cnt) DP
# ---------------------------------------------------------------------------
def min_operations_optimal(s: str) -> int:
    """
    Optimal solution using a left-to-right DP for each candidate target t.

    For fixed t, dp_prev maps {0, 1} -> (best_cost_so_far, excess_of_prev)
    where the key is whether the *previous* letter was chosen active.
    We only need the previous letter's resulting excess (not deficit,
    since deficit can only be "consumed" by the letter *before* it, and
    that's exactly what we handle when we move to the next index).

    Transition for letter i, choice in {0 (exclude), 1 (include @ t)}:
        diff    = cnt[i] - (t if choice else 0)
        excess  = max(diff, 0)
        deficit = max(-diff, 0)
        for each previous choice:
            saving = min(prev_excess, deficit)
            candidate_cost = prev_cost + excess + deficit - saving
        dp_cur[choice] = (min candidate_cost, excess)

    Answer for this t = min(dp_final[0].cost, dp_final[1].cost).
    Overall answer = min over t in [1, max_cnt] of the above
    (t=0 / delete-everything is implicitly reachable and never beats a
    positive-cost baseline in this formulation once we seed best=len(s)).
    """
    n = len(s)
    if n == 0:
        return 0

    count = _counts(s)
    max_count = max(count)
    best = n  # baseline: delete every character (empty string is "good")

    for target in range(1, max_count + 1):
        # dp_prev[choice] = (cost, excess) ; before letter 0 there is no
        # neighbor to save with, so both "virtual previous" states are
        # (cost=0, excess=0).
        dp_prev: Dict[int, Tuple[int, int]] = {0: (0, 0), 1: (0, 0)}

        for i in range(ALPHABET_SIZE):
            count_i = count[i]
            dp_cur: Dict[int, Tuple[int, int]] = {}
            for choice in (0, 1):
                target_choice = target if choice else 0
                diff = count_i - target_choice
                excess = diff if diff > 0 else 0
                deficit = -diff if diff < 0 else 0

                best_cost = None
                for prev_cost, prev_excess in dp_prev.values():
                    saving = min(prev_excess, deficit)
                    cand = prev_cost + excess + deficit - saving
                    if best_cost is None or cand < best_cost:
                        best_cost = cand
                dp_cur[choice] = (best_cost, excess)
            dp_prev = dp_cur

        t_best = min(cost for cost, _ in dp_prev.values())
        if t_best < best:
            best = t_best

    return best


# ---------------------------------------------------------------------------
# Tests
# ---------------------------------------------------------------------------
def _run_examples() -> None:
    examples = [
        ("acab", 1),
        ("wddw", 0),
        ("aaabc", 2),
    ]
    for s, expected in examples:
        bf = min_operations_bruteforce(s)
        opt = min_operations_optimal(s)
        status_bf = "OK" if bf == expected else "FAIL"
        status_opt = "OK" if opt == expected else "FAIL"
        print(f"s={s!r:8} expected={expected}  bruteforce={bf} [{status_bf}]  optimal={opt} [{status_opt}]")


def _stress_test(trials: int = 300, max_len: int = 9, alphabet: str = "abcde") -> None:
    """
    Randomized cross-validation: brute force vs optimal on small strings
    (kept small / low-alphabet-diversity because the brute force is
    exponential in distinct-letter count).
    """
    import random

    random.seed(42)
    mismatches = 0
    for trial in range(trials):
        length = random.randint(1, max_len)
        s = "".join(random.choice(alphabet) for _ in range(length))
        bf = min_operations_bruteforce(s)
        opt = min_operations_optimal(s)
        if bf != opt:
            mismatches += 1
            print(f"MISMATCH on s={s!r}: bruteforce={bf}, optimal={opt}")

    if mismatches == 0:
        print(f"Stress test passed: {trials}/{trials} random strings agree.")
    else:
        print(f"Stress test FAILED: {mismatches}/{trials} mismatches.")


if __name__ == "__main__":
    print("=== Official examples ===")
    _run_examples()
    print("\n=== Randomized stress test (bruteforce vs optimal) ===")
    _stress_test()
