"""
LeetCode Problem Set — Brute Force vs. Optimal Solutions
===========================================================
Covers: House Robber II, K-diff Pairs in an Array, Majority Element,
        Search a 2D Matrix, Decode String, Rotting Oranges

Each problem has:
  - A brute-force (or naive) solution, clearly correct but not efficient
  - An optimal solution, with the technique named
  - Time/space complexity for both
  - A __main__ block with test cases you can run directly:  python leetcode_solutions.py
"""

from collections import deque, defaultdict
from typing import List


# =====================================================================
# 1. HOUSE ROBBER II  (Medium)  — Array / Dynamic Programming
# ---------------------------------------------------------------------
# Houses are arranged in a CIRCLE. Adjacent houses can't both be robbed.
# Maximize total money robbed.
#
# Key insight: because house[0] and house[n-1] are adjacent in a circle,
# you can never rob both. So the answer is the MAX of two linear
# House-Robber-I sub-problems:
#   (a) rob houses [0 .. n-2]  (exclude last house)
#   (b) rob houses [1 .. n-1]  (exclude first house)
# =====================================================================

class HouseRobberII:

    @staticmethod
    def brute_force(nums: List[int]) -> int:
        """
        Pure exhaustive recursion — try robbing or skipping each house,
        with a wraparound check comparing to whichever house we started at.
        No memoization -> exponential blowup, but easiest to reason about
        correctness first.

        Time:  O(2^n)   Space: O(n) recursion stack
        """
        n = len(nums)
        if n == 0:
            return 0
        if n == 1:
            return nums[0]

        def dfs(i: int, first_robbed: bool, can_rob_last: bool) -> int:
            if i >= n:
                return 0
            # Skip this house
            best = dfs(i + 1, first_robbed, can_rob_last)
            # Rob this house, if allowed (not the wraparound-forbidden last house
            # when we robbed the first one)
            if not (i == n - 1 and first_robbed):
                robbed = nums[i] + dfs(i + 2, first_robbed, can_rob_last)
                best = max(best, robbed)
            return best

        # Branch 1: we rob house 0 -> last house is forbidden
        rob_first = nums[0] + dfs(2, True, False)
        # Branch 2: we do NOT rob house 0 -> no wraparound restriction
        skip_first = dfs(1, False, True)
        return max(rob_first, skip_first)

    @staticmethod
    def optimal(nums: List[int]) -> int:
        """
        Linear House Robber (DP, O(1) space) run twice on the two
        non-wrapping sub-arrays, then take the max.

        Time:  O(n)   Space: O(1)
        """
        n = len(nums)
        if n == 0:
            return 0
        if n == 1:
            return nums[0]

        def rob_linear(houses: List[int]) -> int:
            prev, curr = 0, 0
            for money in houses:
                prev, curr = curr, max(curr, prev + money)
            return curr

        return max(rob_linear(nums[:-1]), rob_linear(nums[1:]))


# =====================================================================
# 2. K-DIFF PAIRS IN AN ARRAY  (Medium)  — Array / Hash Map
# ---------------------------------------------------------------------
# Count the number of UNIQUE pairs (i, j) such that |nums[i]-nums[j]| == k.
# k >= 0. Pairs are counted by VALUE, not by index.
# =====================================================================

class KDiffPairs:

    @staticmethod
    def brute_force(nums: List[int], k: int) -> int:
        """
        Check every pair of indices, collect valid pairs (as sorted value
        tuples) in a set to dedupe.

        Time:  O(n^2)   Space: O(n) for the result set
        """
        n = len(nums)
        pairs = set()
        for i in range(n):
            for j in range(i + 1, n):
                if abs(nums[i] - nums[j]) == k:
                    pairs.add((min(nums[i], nums[j]), max(nums[i], nums[j])))
        return len(pairs)

    @staticmethod
    def optimal(nums: List[int], k: int) -> int:
        """
        Frequency map, single pass, O(1) lookups.
          - k == 0: count values that appear 2+ times (each contributes one pair)
          - k  > 0: for each unique value v, check if v + k exists
          - k  < 0: invalid per problem constraints, but guarded anyway

        Time:  O(n)   Space: O(n)
        """
        if k < 0:
            return 0

        freq = defaultdict(int)
        for num in nums:
            freq[num] += 1

        count = 0
        if k == 0:
            count = sum(1 for v in freq.values() if v >= 2)
        else:
            for value in freq:
                if value + k in freq:
                    count += 1
        return count


# =====================================================================
# 3. MAJORITY ELEMENT  (Easy)  — Array / Hash Map / Boyer-Moore Voting
# ---------------------------------------------------------------------
# Find the element that appears more than floor(n / 2) times.
# Guaranteed to exist (problem constraint).
# =====================================================================

class MajorityElement:

    @staticmethod
    def brute_force(nums: List[int]) -> int:
        """
        For each element, count its occurrences across the whole array.
        Stop as soon as a majority is confirmed.

        Time:  O(n^2)   Space: O(1)
        """
        n = len(nums)
        for candidate in nums:
            count = sum(1 for num in nums if num == candidate)
            if count > n // 2:
                return candidate
        return -1  # unreachable given problem guarantee

    @staticmethod
    def optimal(nums: List[int]) -> int:
        """
        Boyer-Moore Voting Algorithm. Maintain a running "candidate" and a
        counter. Incrementing on a match, decrementing on a mismatch; when
        the counter hits 0, swap in a new candidate. The majority element
        (by definition, >n/2 occurrences) always survives this cancellation.

        Time:  O(n)   Space: O(1)
        """
        candidate, count = None, 0
        for num in nums:
            if count == 0:
                candidate = num
            count += 1 if num == candidate else -1
        return candidate


# =====================================================================
# 4. SEARCH A 2D MATRIX  (Medium)  — Binary Search
# ---------------------------------------------------------------------
# Matrix rows are sorted left-to-right; the first integer of each row is
# greater than the last integer of the previous row (so the whole matrix
# is effectively one big sorted 1D array laid out row by row).
# Return True if `target` exists.
# =====================================================================

class Search2DMatrix:

    @staticmethod
    def brute_force(matrix: List[List[int]], target: int) -> bool:
        """
        Scan every cell. Ignores the sorted structure entirely.

        Time:  O(m * n)   Space: O(1)
        """
        for row in matrix:
            for val in row:
                if val == target:
                    return True
        return False

    @staticmethod
    def optimal(matrix: List[List[int]], target: int) -> bool:
        """
        Treat the matrix as a flattened sorted array of length m*n and
        binary search it directly, converting mid -> (row, col) via
        divmod. No need to materialize the flattened array.

        Time:  O(log(m*n))   Space: O(1)
        """
        if not matrix or not matrix[0]:
            return False

        rows, cols = len(matrix), len(matrix[0])
        lo, hi = 0, rows * cols - 1

        while lo <= hi:
            mid = (lo + hi) // 2
            r, c = divmod(mid, cols)
            val = matrix[r][c]
            if val == target:
                return True
            elif val < target:
                lo = mid + 1
            else:
                hi = mid - 1
        return False


# =====================================================================
# 5. DECODE STRING  (Medium)  — String / Stack
# ---------------------------------------------------------------------
# Decode a string like "3[a2[c]]" -> "accaccacc"
# Pattern: k[encoded_string] means encoded_string repeated k times,
# and these can nest arbitrarily.
# =====================================================================

class DecodeString:

    @staticmethod
    def brute_force(s: str) -> str:
        """
        Recursive-descent parser using a mutable index captured via a list
        (Python has no easy pass-by-reference int). Conceptually simple —
        directly mirrors the grammar — but relies on the call stack
        rather than an explicit one, so it's the "naive" approach here
        even though its complexity is comparable to the iterative version.

        Time:  O(maxK^countK * n) worst case string blow-up dominates both
               approaches; here we flag it as "brute" because recursion
               depth grows with nesting depth and it recomputes substring
               concatenation with += in a way that can be O(n^2) for
               pathological inputs.
        Space: O(n) recursion stack + output
        """
        pos = [0]  # mutable box so the nested function can advance it

        def decode() -> str:
            result = ""
            while pos[0] < len(s) and s[pos[0]] != ']':
                if s[pos[0]].isdigit():
                    num = 0
                    while s[pos[0]].isdigit():
                        num = num * 10 + int(s[pos[0]])
                        pos[0] += 1
                    pos[0] += 1  # skip '['
                    inner = decode()
                    pos[0] += 1  # skip ']'
                    result += inner * num
                else:
                    result += s[pos[0]]
                    pos[0] += 1
            return result

        return decode()

    @staticmethod
    def optimal(s: str) -> str:
        """
        Iterative, explicit-stack approach. Push (current_string, repeat_count)
        onto the stack whenever we hit '[', and pop-and-combine on ']'.
        This is the standard "optimal" pattern for this problem: avoids
        recursion-depth limits entirely and is easy to trace by hand.

        Time:  O(maxK^countK * n) — dominated by the size of the fully
               decoded output, same theoretical bound as any correct
               solution; the iterative version just avoids call-stack
               overhead and recursion-depth limits.
        Space: O(n) for the stack
        """
        stack = []          # list of (previous_string, repeat_count)
        current = ""
        num = 0

        for ch in s:
            if ch.isdigit():
                num = num * 10 + int(ch)
            elif ch == '[':
                stack.append((current, num))
                current = ""
                num = 0
            elif ch == ']':
                prev_string, repeat_count = stack.pop()
                current = prev_string + current * repeat_count
            else:
                current += ch

        return current


# =====================================================================
# 6. ROTTING ORANGES  (Medium)  — Grid / Multi-source BFS
# ---------------------------------------------------------------------
# Grid: 0 = empty, 1 = fresh orange, 2 = rotten orange.
# Every minute, any fresh orange adjacent (4-directionally) to a rotten
# one becomes rotten. Return minutes until no fresh orange remains,
# or -1 if impossible.
# =====================================================================

class RottingOranges:

    @staticmethod
    def brute_force(grid: List[List[int]]) -> int:
        """
        Simulate minute-by-minute: on each pass, scan the ENTIRE grid and
        rot any fresh orange adjacent to a rotten one THAT minute. Repeat
        until a full pass makes no changes. This re-scans the whole grid
        every minute even though only the "frontier" oranges matter.

        Time:  O((m*n)^2) worst case — up to m*n minutes, each requiring
               a full m*n scan.
        Space: O(1) extra (mutates grid in place; ignoring input storage)
        """
        if not grid or not grid[0]:
            return 0
        rows, cols = len(grid), len(grid[0])

        def has_fresh() -> bool:
            return any(cell == 1 for row in grid for cell in row)

        minutes = 0
        while has_fresh():
            newly_rotten = []
            for r in range(rows):
                for c in range(cols):
                    if grid[r][c] == 2:
                        for dr, dc in ((1, 0), (-1, 0), (0, 1), (0, -1)):
                            nr, nc = r + dr, c + dc
                            if 0 <= nr < rows and 0 <= nc < cols and grid[nr][nc] == 1:
                                newly_rotten.append((nr, nc))
            if not newly_rotten:
                return -1  # fresh oranges remain but nothing can rot them
            for r, c in newly_rotten:
                grid[r][c] = 2
            minutes += 1
        return minutes

    @staticmethod
    def optimal(grid: List[List[int]]) -> int:
        """
        Multi-source BFS. Seed the queue with ALL initially-rotten oranges
        at once (minute 0), then expand level-by-level (each BFS "layer"
        = one minute). This is the textbook-optimal pattern for
        "simultaneous spread" grid problems.

        Time:  O(m * n)  — every cell is enqueued/processed at most once
        Space: O(m * n)  — queue in the worst case (all oranges rotten)
        """
        if not grid or not grid[0]:
            return 0
        rows, cols = len(grid), len(grid[0])

        queue = deque()
        fresh_count = 0

        for r in range(rows):
            for c in range(cols):
                if grid[r][c] == 2:
                    queue.append((r, c, 0))  # (row, col, minute_rotted)
                elif grid[r][c] == 1:
                    fresh_count += 1

        if fresh_count == 0:
            return 0

        minutes_elapsed = 0
        while queue:
            r, c, minute = queue.popleft()
            minutes_elapsed = max(minutes_elapsed, minute)
            for dr, dc in ((1, 0), (-1, 0), (0, 1), (0, -1)):
                nr, nc = r + dr, c + dc
                if 0 <= nr < rows and 0 <= nc < cols and grid[nr][nc] == 1:
                    grid[nr][nc] = 2
                    fresh_count -= 1
                    queue.append((nr, nc, minute + 1))

        return minutes_elapsed if fresh_count == 0 else -1


# =====================================================================
# TEST / DEMO
# =====================================================================

if __name__ == "__main__":

    print("=== 1. House Robber II ===")
    for nums in [[2, 3, 2], [1, 2, 3, 1], [1, 2, 3], [200], [1, 2]]:
        bf = HouseRobberII.brute_force(nums)
        opt = HouseRobberII.optimal(nums)
        print(f"  nums={nums!r:20} brute={bf}  optimal={opt}  match={bf == opt}")

    print("\n=== 2. K-diff Pairs in an Array ===")
    cases = [
        ([3, 1, 4, 1, 5], 2),
        ([1, 2, 3, 4, 5], 1),
        ([1, 3, 1, 5, 4], 0),
        ([1, 1, 1, 1, 1], 0),
        ([1, 2, 4, 4, 3, 3, 0, 9, 2, 3], 3),
    ]
    for nums, k in cases:
        bf = KDiffPairs.brute_force(nums, k)
        opt = KDiffPairs.optimal(nums, k)
        print(f"  nums={nums!r:35} k={k}  brute={bf}  optimal={opt}  match={bf == opt}")

    print("\n=== 3. Majority Element ===")
    for nums in [[3, 2, 3], [2, 2, 1, 1, 1, 2, 2], [1], [6, 5, 5]]:
        bf = MajorityElement.brute_force(nums)
        opt = MajorityElement.optimal(nums)
        print(f"  nums={nums!r:25} brute={bf}  optimal={opt}  match={bf == opt}")

    print("\n=== 4. Search a 2D Matrix ===")
    matrix = [[1, 3, 5, 7], [10, 11, 16, 20], [23, 30, 34, 60]]
    for target in [3, 13, 60, 1, -5]:
        bf = Search2DMatrix.brute_force(matrix, target)
        opt = Search2DMatrix.optimal(matrix, target)
        print(f"  target={target:4} brute={bf}  optimal={opt}  match={bf == opt}")

    print("\n=== 5. Decode String ===")
    for s in ["3[a]2[bc]", "3[a2[c]]", "2[abc]3[cd]ef", "abc", "10[leetcode]"]:
        bf = DecodeString.brute_force(s)
        opt = DecodeString.optimal(s)
        print(f"  s={s!r:20} brute={bf!r:35} optimal={opt!r:35} match={bf == opt}")

    print("\n=== 6. Rotting Oranges ===")
    grids = [
        [[2, 1, 1], [1, 1, 0], [0, 1, 1]],
        [[2, 1, 1], [0, 1, 1], [1, 0, 1]],
        [[0, 2]],
        [[0]],
        [[1]],
    ]
    for g in grids:
        g1 = [row[:] for row in g]
        g2 = [row[:] for row in g]
        bf = RottingOranges.brute_force(g1)
        opt = RottingOranges.optimal(g2)
        print(f"  grid={g!r:40} brute={bf}  optimal={opt}  match={bf == opt}")
