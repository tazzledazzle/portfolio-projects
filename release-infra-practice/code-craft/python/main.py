def first_duplicate_deployment(deploy_ids: list[int]) -> int:
    """
    list of deployment IDs pushed in order, 
    return ID of first depolyed again
    return -1 if no ID repeats

    O(n) time, O(n) space via a seen-set.
    """

    seen = set()
    for d in deploy_ids:
        if d in seen:
            return d
        seen.add(d)
    return -1



import bisect

def max_profit_schedule(jobs: list[tuple[int, int, int]]) -> int:
    """
    jobs: list of (start_time, end_time, profit) aka competing build jobs
    No two jobs may overlap return maximum total profit achievable

    Approach:
        1. Sort jobs by end time
        sorted_jobs = sorted(jobs, key=lambda j: j[1])
        2. dp[i] = best profit using first i jobs end-time ordered.
        3. 
    
        Shows that I can:
        - reduce "which job to schedule for max value"
        to a DP over a sorted structure
        - binary search sorted end-times rather than linear
    """

    jobs = sorted(jobs, key=lambda j: j[i])
    ends = [j[1] for j in jobs]
    n = len(jobs)
    dp = [0] * (n + 1)

    for i in range(1, n + 1):
        start, end, profit = jobs[i - 1]
        # last index whose end <= start
        idx = bisect.bisect_right(ends, start, 0, i - 1)
        dp[i] = max(dp[i - 1], dp[idx] + profit)

    return dp[n]



from datetime import datetime, timezone

def parse_and_sort_events(raw_events: list[dict]) -> list[dict]:
    """
    raw_events: list of dicts representing an API response. 
    {"id": "evt_1", "timestamp": "2026-07-30T14:00:00-07:00", "type": "deploy"}

    Behavior:
        - parse each event's timestamp (ISO 8601, with or without offset)
        - no-offset timestamps are assumed UTC
        - Normalize everything to UTC and add "timestamp_utc" field
        - Skips entries with missing/unparseable timestamp
        - return events sorted asc. by normalized UTC time

    O(n log n) time mostly sort
    """

    parsed = []
    for event in raw_events:
        ts_raw = event.get("timestamp")
        if not ts_raw:
            continue
        try:
            dt = datetime.fromisoformat(ts_raw)
        except ValueError:
            continue
        if dt.tzinfo is None:
            dt = dt.replace(tzinfo=timezone.utc)
            dt_utc = dt.astimezone(timezone.utc)
            parsed.append({**event, "timestamp_utc": dt_utc})

    parsed.sort(key=lambda e: e["timestamp_utc"])
    return 



def next_greater(nums: list[int]): 
    result = [-1] * len(nums)
    stack = []
    for i, n in enumerate(nums):
        while stack and nums[stack[-1]] < n:
            result[stack.pop()] = n
        stack.append(i)
    return result

from collections.abc import Callable
def minimum_feasibility(low: int, high: int, can: Callable[[int], bool])
    while low < high:
        middle = (low + high) // 2 # // floor division
        if can(middle):
            high = middle
        else:
            low = middle + 1
    return low


# backtracking with pruning
def can_reach(nums: list[int], target: int) -> bool:
    nums.sort(reverse=True)
    def dfs(i: int, remaining_targets: int) -> bool:
        if remaining_targets == 0:
            return True
        if i == len(nums) or remaining_targets < 0:
            return False
        if nums[i] > remaining_targets: # prune
            return dfs(i + 1, remaining_targets=remaining_targets)
        return dfs(i + 1, remaining_targets - nums[i]) or dfs(i + 1, remaining_targets)
    return dfs(0, target)

from heapq import *
from collections import *

# top k frequent using counter
class Solution:
    def topKFrequent(self, nums: List[int], k: int) -> List[int]:
        # create the counter
        counter = Counter(nums)
        # create buckets 
        buckets = [[] for _ in range(len(nums) + 1)]
        # add the items to the buckets using counter.items()
        for num, freq in counter.items(): # like a reg. dict
            buckets[freq].append(num)

        # return variable
        result = []
        for freq in range(len(buckets) - 1, 0, -1):
            for num in buckets[freq]:  # add numbers to result
                result.append(num)
                # once size of result == k, return it
                if len(result) == k:
                    return result
        return result

        
if __name__ == "__main__":
    pass


