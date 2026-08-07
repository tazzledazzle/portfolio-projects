"""
Q9 — Debug and Explain Fixes

This module contains intentionally buggy implementations. Your task:
1. Identify each bug (document in the docstring below each function).
2. Fix it — the tests will pass when all bugs are corrected.

Bugs are purposely subtle: off-by-one, mutating shared state, wrong comparison
operator, missing edge-case handling, etc.
"""

from __future__ import annotations

from decimal import Decimal


# BUG 1: wrong boundary — should allow purchases exactly at the limit
def can_purchase(balance: Decimal, limit: Decimal, amount: Decimal) -> bool:
    """Return True if (balance + amount) < limit."""
    return balance + amount < limit  # TODO: fix the comparison


# BUG 2: accumulates across calls due to mutable default argument
def collect_transaction_ids(txn_id: str, seen: list[str] = []) -> list[str]:  # noqa: B006
    """Append txn_id to seen and return the list."""
    seen.append(txn_id)
    return seen  # TODO: fix mutable default argument


# BUG 3: integer division truncates points — should use Decimal arithmetic
def compute_points(usd_cents: int) -> int:
    """Return points where 100 cents = 150 points (1.5x rate)."""
    return usd_cents * 3 // 2  # TODO: this loses precision for odd amounts


# BUG 4: iterates over a copy so removals don't take effect
def remove_expired(transactions: list[dict[str, object]], cutoff_ts: float) -> list[dict[str, object]]:
    """Remove transactions older than cutoff_ts in-place and return the list."""
    for txn in transactions[:]:  # BUG: should modify original, not copy
        if txn["timestamp"] < cutoff_ts:  # type: ignore[operator]
            transactions.remove(txn)
    return transactions  # TODO: verify in-place removal is correct


# BUG 5: silent swallow of ValueError makes bad input indistinguishable from zero
def safe_parse_amount(raw: str) -> Decimal:
    """Parse a monetary string; return 0 on any error."""
    try:
        return Decimal(raw)
    except Exception:
        return Decimal("0")  # TODO: should raise ValueError for non-numeric input
