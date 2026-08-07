"""Tests for Q9 — Debug and Explain Fixes (all should FAIL until bugs are fixed)."""

from decimal import Decimal

import pytest

from q09_debug_and_fix import (
    can_purchase,
    collect_transaction_ids,
    compute_points,
    remove_expired,
    safe_parse_amount,
)


# --- BUG 1: can_purchase boundary ---

def test_can_purchase_within_limit() -> None:
    assert can_purchase(Decimal("0"), Decimal("1000"), Decimal("500")) is True


def test_can_purchase_exact_limit_allowed() -> None:
    # Should be True: balance + amount == limit is allowed
    assert can_purchase(Decimal("0"), Decimal("1000"), Decimal("1000")) is True


def test_can_purchase_over_limit_rejected() -> None:
    assert can_purchase(Decimal("0"), Decimal("1000"), Decimal("1001")) is False


# --- BUG 2: mutable default argument ---

def test_collect_txn_ids_isolated_calls() -> None:
    result1 = collect_transaction_ids("t1")
    result2 = collect_transaction_ids("t2")
    assert "t1" not in result2


def test_collect_txn_ids_returns_correct_list() -> None:
    result = collect_transaction_ids("tx-abc", [])
    assert result == ["tx-abc"]


# --- BUG 3: integer division precision ---

def test_compute_points_even_cents() -> None:
    assert compute_points(200) == 300  # 200 * 1.5 = 300


def test_compute_points_odd_cents() -> None:
    # 101 cents * 1.5 = 151.5 → should round correctly (not truncate to 151)
    assert compute_points(101) >= 151


def test_compute_points_one_cent() -> None:
    # 1 cent → at least 1 point (not 0 due to truncation)
    assert compute_points(1) >= 1


# --- BUG 4: remove_expired modifies in-place ---

def test_remove_expired_removes_old() -> None:
    txns: list[dict[str, object]] = [
        {"id": "t1", "timestamp": 100.0},
        {"id": "t2", "timestamp": 200.0},
    ]
    result = remove_expired(txns, cutoff_ts=150.0)
    assert all(t["timestamp"] >= 150.0 for t in result)  # type: ignore[operator]


def test_remove_expired_modifies_original_list() -> None:
    txns: list[dict[str, object]] = [{"id": "old", "timestamp": 50.0}]
    remove_expired(txns, cutoff_ts=100.0)
    assert len(txns) == 0


def test_remove_expired_keeps_recent() -> None:
    txns: list[dict[str, object]] = [{"id": "new", "timestamp": 200.0}]
    remove_expired(txns, cutoff_ts=100.0)
    assert len(txns) == 1


# --- BUG 5: safe_parse_amount ---

def test_safe_parse_amount_valid() -> None:
    assert safe_parse_amount("99.99") == Decimal("99.99")


def test_safe_parse_amount_invalid_raises() -> None:
    with pytest.raises(ValueError):
        safe_parse_amount("NOT_A_NUMBER")


def test_safe_parse_amount_empty_raises() -> None:
    with pytest.raises(ValueError):
        safe_parse_amount("")
