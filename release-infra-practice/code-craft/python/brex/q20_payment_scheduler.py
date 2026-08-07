"""
Q20 — Recurring/Subscription Payment Scheduler

Base: given a schedule (anchor date + interval), generate the next N due dates.
Ext 1: process a due payment; handle a missed cycle by detecting gaps and retrying.
Ext 2: full retry logic — if a payment fails, schedule a retry within the same window.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from datetime import date, timedelta
from decimal import Decimal
from enum import Enum
from typing import Callable


class Interval(Enum):
    WEEKLY = 7
    BIWEEKLY = 14
    MONTHLY = 30  # simplified: 30-day months


class PaymentOutcome(Enum):
    SUCCESS = "success"
    FAILED = "failed"
    SKIPPED = "skipped"


@dataclass
class Schedule:
    schedule_id: str
    subscriber_id: str
    amount: Decimal
    interval: Interval
    anchor_date: date
    last_processed: date | None = None


@dataclass
class PaymentAttempt:
    schedule_id: str
    due_date: date
    outcome: PaymentOutcome
    retry_count: int = 0


def next_due_dates(schedule: Schedule, n: int, *, from_date: date | None = None) -> list[date]:
    """
    Return the next n due dates for schedule, starting after from_date
    (defaults to schedule.last_processed or anchor_date - 1 day).
    """
    raise NotImplementedError


def detect_missed_cycles(schedule: Schedule, as_of: date) -> list[date]:
    """
    Return any due dates that passed between last_processed and as_of
    that were never processed.
    """
    raise NotImplementedError


def process_payment(
    schedule: Schedule,
    due_date: date,
    charge_fn: Callable[[str, Decimal], bool],
    *,
    max_retries: int = 3,
) -> PaymentAttempt:
    """
    Call charge_fn(subscriber_id, amount). Retry up to max_retries times on failure.
    Record outcome and retry count. Update schedule.last_processed on success.
    """
    raise NotImplementedError
