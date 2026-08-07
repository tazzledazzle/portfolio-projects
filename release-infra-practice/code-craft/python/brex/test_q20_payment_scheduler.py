"""Tests for Q20 — Recurring/Subscription Payment Scheduler."""

from datetime import date
from decimal import Decimal

import pytest

from q20_payment_scheduler import (
    Interval,
    PaymentAttempt,
    PaymentOutcome,
    Schedule,
    detect_missed_cycles,
    next_due_dates,
    process_payment,
)

ANCHOR = date(2024, 1, 1)


def make_schedule(**kwargs: object) -> Schedule:
    defaults: dict[str, object] = {
        "schedule_id": "s1",
        "subscriber_id": "sub-1",
        "amount": Decimal("29.99"),
        "interval": Interval.MONTHLY,
        "anchor_date": ANCHOR,
    }
    defaults.update(kwargs)
    return Schedule(**defaults)  # type: ignore[arg-type]


def test_next_due_dates_returns_n_dates() -> None:
    s = make_schedule()
    dates = next_due_dates(s, 3)
    assert len(dates) == 3


def test_next_due_dates_weekly_interval() -> None:
    s = make_schedule(interval=Interval.WEEKLY)
    dates = next_due_dates(s, 2)
    assert (dates[1] - dates[0]).days == 7


def test_next_due_dates_monthly_interval() -> None:
    s = make_schedule()
    dates = next_due_dates(s, 2)
    assert (dates[1] - dates[0]).days == 30


def test_next_due_dates_starts_after_anchor() -> None:
    s = make_schedule()
    dates = next_due_dates(s, 1)
    assert dates[0] > ANCHOR or dates[0] == ANCHOR


def test_next_due_dates_respects_from_date() -> None:
    s = make_schedule()
    from_d = date(2024, 3, 1)
    dates = next_due_dates(s, 1, from_date=from_d)
    assert dates[0] > from_d


def test_detect_missed_no_missed_when_current() -> None:
    s = make_schedule(last_processed=ANCHOR)
    missed = detect_missed_cycles(s, as_of=date(2024, 1, 20))
    assert missed == []


def test_detect_missed_one_cycle() -> None:
    s = make_schedule(last_processed=ANCHOR)
    missed = detect_missed_cycles(s, as_of=date(2024, 2, 15))
    assert len(missed) == 1


def test_detect_missed_multiple_cycles() -> None:
    s = make_schedule(last_processed=ANCHOR)
    missed = detect_missed_cycles(s, as_of=date(2024, 4, 15))
    assert len(missed) >= 2


def test_process_payment_success_updates_last_processed() -> None:
    s = make_schedule()
    process_payment(s, date(2024, 2, 1), lambda _id, _amt: True)
    assert s.last_processed == date(2024, 2, 1)


def test_process_payment_success_outcome() -> None:
    s = make_schedule()
    attempt = process_payment(s, date(2024, 2, 1), lambda _id, _amt: True)
    assert attempt.outcome == PaymentOutcome.SUCCESS


def test_process_payment_failure_outcome() -> None:
    s = make_schedule()
    attempt = process_payment(s, date(2024, 2, 1), lambda _id, _amt: False, max_retries=1)
    assert attempt.outcome == PaymentOutcome.FAILED


def test_process_payment_retries_on_failure() -> None:
    calls: list[int] = []
    def charge(_id: str, _amt: Decimal) -> bool:
        calls.append(1)
        return False
    s = make_schedule()
    process_payment(s, date(2024, 2, 1), charge, max_retries=3)
    assert len(calls) == 3


def test_process_payment_retry_count_recorded() -> None:
    s = make_schedule()
    attempt = process_payment(s, date(2024, 2, 1), lambda _id, _amt: False, max_retries=3)
    assert attempt.retry_count == 3


def test_process_payment_no_retry_on_success() -> None:
    calls: list[int] = []
    def charge(_id: str, _amt: Decimal) -> bool:
        calls.append(1)
        return True
    s = make_schedule()
    process_payment(s, date(2024, 2, 1), charge, max_retries=3)
    assert len(calls) == 1
