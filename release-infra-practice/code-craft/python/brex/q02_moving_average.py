"""
Q2 — Moving Average of Transaction Volumes

Base: maintain a sliding window of size k over a stream of transaction volume
      values and return the current moving average after each insertion.
Ext 1: support querying the moving average at any point (not just after insert).
Ext 2: support resetting the window and tracking the global running total.
"""

from __future__ import annotations

from collections import deque
from decimal import Decimal


class MovingAverage:
    """Sliding-window moving average over a numeric stream."""

    def __init__(self, window_size: int) -> None:
        if window_size <= 0:
            raise ValueError("window_size must be positive")
        self.window_size = window_size
        self._window: deque[Decimal] = deque()
        self._running_total: Decimal = Decimal("0")
        self._global_total: Decimal = Decimal("0")

    def add(self, value: Decimal) -> Decimal:
        """Insert a value and return the current moving average."""
        if len(self._window) == self.window_size:
            item = self._window.popleft()
            self._running_total -= item
        self._window.append(value)
        self._running_total += value
        self._global_total += value

        return self.current_average

    @property
    def current_average(self) -> Decimal:
        """Return the moving average of values currently in the window."""
        return self._running_total / len(self._window)

    @property
    def global_total(self) -> Decimal:
        """Return the sum of all values ever inserted (not just the window)."""
        return self._global_total

    def reset(self) -> None:
        """Clear the window. global_total is preserved."""
        self._window.clear()
        self._running_total = Decimal("0")
