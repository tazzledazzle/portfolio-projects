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

if __name__ == "__main__":
    pass


