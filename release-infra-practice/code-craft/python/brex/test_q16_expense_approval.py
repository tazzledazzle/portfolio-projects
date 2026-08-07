"""Tests for Q16 — Expense Auto-Approval Flow."""

from decimal import Decimal

import pytest

from q16_expense_approval import ApprovalQueue, ApprovalStatus, Expense, categorize, submit_expense


def make_expense(expense_id: str = "e1", amount: str = "50", mcc: str = "5812") -> Expense:
    return Expense(expense_id=expense_id, amount=Decimal(amount), mcc=mcc, description="test")


def test_categorize_known_mcc() -> None:
    assert categorize("5812") == "dining"


def test_categorize_unknown_mcc_returns_other() -> None:
    assert categorize("9999") == "other"


def test_submit_expense_under_threshold_auto_approved() -> None:
    exp = make_expense(amount="50")
    status = submit_expense(exp)
    assert status == ApprovalStatus.AUTO_APPROVED


def test_submit_expense_sets_category() -> None:
    exp = make_expense(mcc="7011")
    submit_expense(exp)
    assert exp.category == "lodging"


def test_submit_expense_over_threshold_pending_manager() -> None:
    exp = make_expense(amount="500")
    status = submit_expense(exp)
    assert status == ApprovalStatus.PENDING_MANAGER


def test_submit_expense_at_threshold_auto_approved() -> None:
    exp = make_expense(amount="100")
    status = submit_expense(exp)
    assert status == ApprovalStatus.AUTO_APPROVED


def test_queue_enqueue_pending_expense() -> None:
    q = ApprovalQueue()
    exp = make_expense(amount="500")
    submit_expense(exp)
    q.enqueue(exp)
    assert len(q.pending) == 1


def test_queue_enqueue_non_pending_raises() -> None:
    q = ApprovalQueue()
    exp = make_expense(amount="50")
    submit_expense(exp)
    with pytest.raises(ValueError):
        q.enqueue(exp)


def test_queue_approve_sets_status() -> None:
    q = ApprovalQueue()
    exp = make_expense(amount="500")
    submit_expense(exp)
    q.enqueue(exp)
    result = q.approve("e1")
    assert result.status == ApprovalStatus.MANAGER_APPROVED


def test_queue_approve_removes_from_pending(q: ApprovalQueue = ApprovalQueue()) -> None:
    q2 = ApprovalQueue()
    exp = make_expense(expense_id="e2", amount="500")
    submit_expense(exp)
    q2.enqueue(exp)
    q2.approve("e2")
    assert len(q2.pending) == 0


def test_queue_reject_sets_status() -> None:
    q = ApprovalQueue()
    exp = make_expense(expense_id="e3", amount="500")
    submit_expense(exp)
    q.enqueue(exp)
    result = q.reject("e3")
    assert result.status == ApprovalStatus.MANAGER_REJECTED


def test_queue_approve_unknown_raises() -> None:
    q = ApprovalQueue()
    with pytest.raises(KeyError):
        q.approve("nonexistent")


def test_queue_multiple_pending() -> None:
    q = ApprovalQueue()
    for i in range(3):
        exp = make_expense(expense_id=f"e{i}", amount="500")
        submit_expense(exp)
        q.enqueue(exp)
    assert len(q.pending) == 3


def test_submit_updates_expense_status_field() -> None:
    exp = make_expense(amount="50")
    submit_expense(exp)
    assert exp.status == ApprovalStatus.AUTO_APPROVED


def test_categorize_software_mcc() -> None:
    assert categorize("7372") == "software"
