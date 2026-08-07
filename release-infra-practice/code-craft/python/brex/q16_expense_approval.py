"""
Q16 — Expense Auto-Approval Flow

Base: categorize an incoming expense by merchant category code (MCC).
Ext 1: auto-approve if amount is under a configurable threshold; route to manager otherwise.
Ext 2: manager approval queue — manager can approve or reject queued expenses.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from decimal import Decimal
from enum import Enum

# Simplified MCC → category mapping
MCC_CATEGORIES: dict[str, str] = {
    "5812": "dining",
    "7011": "lodging",
    "4511": "travel",
    "5940": "equipment",
    "7372": "software",
}

AUTO_APPROVE_THRESHOLD: Decimal = Decimal("100")


class ApprovalStatus(Enum):
    AUTO_APPROVED = "auto_approved"
    PENDING_MANAGER = "pending_manager"
    MANAGER_APPROVED = "manager_approved"
    MANAGER_REJECTED = "manager_rejected"


@dataclass
class Expense:
    expense_id: str
    amount: Decimal
    mcc: str
    description: str
    status: ApprovalStatus = ApprovalStatus.PENDING_MANAGER
    category: str = ""


def categorize(mcc: str) -> str:
    """Return the category string for an MCC. Return 'other' if unknown."""
    raise NotImplementedError


def submit_expense(expense: Expense, threshold: Decimal = AUTO_APPROVE_THRESHOLD) -> ApprovalStatus:
    """
    Set expense.category and expense.status based on amount vs threshold.
    Returns the resulting status.
    """
    raise NotImplementedError


class ApprovalQueue:
    def __init__(self) -> None:
        self._queue: dict[str, Expense] = {}

    def enqueue(self, expense: Expense) -> None:
        """Add a PENDING_MANAGER expense to the queue. Raise ValueError if not pending."""
        raise NotImplementedError

    def approve(self, expense_id: str) -> Expense:
        """Mark expense as MANAGER_APPROVED and remove from queue."""
        raise NotImplementedError

    def reject(self, expense_id: str) -> Expense:
        """Mark expense as MANAGER_REJECTED and remove from queue."""
        raise NotImplementedError

    @property
    def pending(self) -> list[Expense]:
        """Return all queued expenses."""
        raise NotImplementedError
