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

# Group anagrams
from collections import *

class Solution:
    def groupAnagrams(self, strs: List[str]) -> List[List[str]]:
        anagram_map = defaultdict(list)
        for st in strs:
            # sort characters for key
            sorted_key = "".join(sorted(st))
           # lean on stdlib
            anagram_map[sorted_key].append(st)
        # return lists as groups
        return list(anagram_map.values())


# Subarray Sum K
def subarraySum(nums: list[int], k: int) -> int:
    # Hash map to store the frequency of prefix sums
    # Base case: a prefix sum of 0 has occurred 1 time (handles subarrays starting at index 0)
    prefix_counts = {0: 1}
    
    current_sum = 0
    total_subarrays = 0
    
    for num in nums:
        # Update the running prefix sum
        current_sum += num
        
        # Check if (current_sum - k) exists in our history
        diff = current_sum - k
        if diff in prefix_counts:
            total_subarrays += prefix_counts[diff]
            
        # Record the current prefix sum in the hash map
        prefix_counts[current_sum] = prefix_counts.get(current_sum, 0) + 1
        
    return total_subarrays
        
if __name__ == "__main__":
    pass


class Solution:
    def isValid(self, s: str) -> bool:
        # create lookup key
        pairs = {")": "(", "]": "[", "}": "{"}
        stack = []
        # for each character in s
        for char in s:
            if char in pairs:
                top_ele = stack.pop() if stack else '#'
                # missmatch
                if pairs[char] != top_ele:
                    return False
            else:
                # into
                stack.append(char)
        # valid if empty
        return len(stack) == 0
                    

def maxProfit(prices: list[int]) -> int:
    # edge cases
    if not prices:
        return 0

    # establish rolling variables
    min_price = float('inf')
    max_profit = 0

    for price in prices:
        # track lowest purchase price seen
        if price < min_price:
            min_price = price
        elif price - min_price > max_profit:
            max_profit = price - min_price
    return max_profit


def isPalindrome(s: str) -> bool:
    left, right = 0, len(s) - 1
    
    while left < right:
        # Skip non-alphanumeric characters from the left
        while left < right and not s[left].isalnum():
            left += 1
        # Skip non-alphanumeric characters from the right
        while left < right and not s[right].isalnum():
            right -= 1
            
        # Compare characters after converting them to lowercase
        if s[left].lower() != s[right].lower():
            return False
            
        left += 1
        right -= 1
        
    return True

# Invert Binary Tree
# Definition for a binary tree node.
# class TreeNode:
#     def __init__(self, val=0, left=None, right=None):
#         self.val = val
#         self.left = left
#         self.right = right
class Solution:
    def invertTree(self, root: Optional[TreeNode]) -> Optional[TreeNode]:
        # base case
        if not root:
            return None

        # Flip
        root.left, root.right = root.right, root.left

        # recurse
        self.invertTree(root.left)
        self.invertTree(root.right)

        # return root
        return root

class Solution:
    def search(self, nums: List[int], target: int) -> int:
        low, high = 0, len(nums) - 1

        while low <= high:
            mid = (low + high ) // 2
            if nums[mid] == target:
                return mid
            elif nums[mid] < target:
                # right side
                low = mid + 1  
            else:
                high = mid - 1
        return -1

# 733. Flood fill
class Solution:
    def floodFill(self, image: List[List[int]], sr: int, sc: int, color: int) -> List[List[int]]:
        original = image[sr][sc]

        if original == color:
            return image
        
        def dfs(row_idx, col_idx):

            if (
                row_idx < 0
                or row_idx >= len(image)
                or col_idx < 0
                or col_idx >= len(image[0])
                or image[row_idx][col_idx] != original
            ): 
                return
            
            image[row_idx][col_idx] = color  # change color

            # traverse all 4 directions (up, down, left, right)
            dfs(row_idx + 1, col_idx)
            dfs(row_idx - 1, col_idx)
            dfs(row_idx, col_idx + 1)
            dfs(row_idx, col_idx - 1)

        dfs(sr, sc)
        return image

# LCA
# Definition for a binary tree node.
# class TreeNode:
#     def __init__(self, x):
#         self.val = x
#         self.left = None
#         self.right = None

class Solution:
    def lowestCommonAncestor(self, root: 'TreeNode', p: 'TreeNode', q: 'TreeNode') -> 'TreeNode':
        if root is None or root == p or root == q:
            return root
        left = self.lowestCommonAncestor(root.left, p, q)
        right = self.lowestCommonAncestor(root.right, p, q)

        if left and right:
            return root
        return left if left else right


# Implement queue using stacks
class MyQueue:

    def __init__(self):
        self.enqueue = []
        self.dequeue = []

    def push(self, x: int) -> None:
        self.enqueue.append(x)
        

    def pop(self) -> int:
        self.peek()
        return self.dequeue.pop()
        

    def peek(self) -> int:
        if not self.dequeue:
            while self.enqueue:
                self.dequeue.append(self.enqueue.pop())
        return self.dequeue[-1]
        

    def empty(self) -> bool:
        return not self.enqueue and not self.dequeue
