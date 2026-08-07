"""Tests for Q10 — Idempotent Card-Transaction Processor."""

from datetime import datetime, timedelta, timezone
from decimal import Decimal

import pytest

from q10_idempotent_card_processor import CardProcessor

T0 = datetime(2024, 1, 1, 12, 0, 0, tzinfo=timezone.utc)


@pytest.fixture
def proc() -> CardProcessor:
    return CardProcessor()


def test_accepted_transaction(proc: CardProcessor) -> None:
    result = proc.process("card-1", "txn-1", Decimal("50"), now=T0)
    assert result == "accepted"


def test_duplicate_key_rejected(proc: CardProcessor) -> None:
    proc.process("card-1", "txn-1", Decimal("50"), now=T0)
    result = proc.process("card-1", "txn-1", Decimal("50"), now=T0)
    assert result == "duplicate"


def test_different_keys_both_accepted(proc: CardProcessor) -> None:
    r1 = proc.process("card-1", "txn-1", Decimal("10"), now=T0)
    r2 = proc.process("card-1", "txn-2", Decimal("10"), now=T0)
    assert r1 == "accepted"
    assert r2 == "accepted"


def test_rate_limit_triggered_at_five(proc: CardProcessor) -> None:
    for i in range(5):
        proc.process("card-1", f"txn-{i}", Decimal("10"), now=T0)
    result = proc.process("card-1", "txn-5", Decimal("10"), now=T0)
    assert result == "rate_limited"


def test_rate_limit_resets_after_window(proc: CardProcessor) -> None:
    for i in range(5):
        proc.process("card-1", f"txn-{i}", Decimal("10"), now=T0)
    later = T0 + timedelta(seconds=61)
    result = proc.process("card-1", "txn-new", Decimal("10"), now=later)
    assert result == "accepted"


def test_is_rate_limited_false_initially(proc: CardProcessor) -> None:
    assert proc.is_rate_limited(now=T0) is False


def test_is_rate_limited_true_at_limit(proc: CardProcessor) -> None:
    for i in range(5):
        proc.process("card-1", f"txn-{i}", Decimal("10"), now=T0)
    assert proc.is_rate_limited(now=T0) is True


def test_accepted_txn_written_to_ledger(proc: CardProcessor) -> None:
    proc.process("card-1", "txn-1", Decimal("75"), now=T0)
    assert len(proc.ledger) == 1


def test_duplicate_not_written_to_ledger(proc: CardProcessor) -> None:
    proc.process("card-1", "txn-1", Decimal("75"), now=T0)
    proc.process("card-1", "txn-1", Decimal("75"), now=T0)
    assert len(proc.ledger) == 1


def test_rate_limited_not_written_to_ledger(proc: CardProcessor) -> None:
    for i in range(5):
        proc.process("card-1", f"txn-{i}", Decimal("10"), now=T0)
    proc.process("card-1", "txn-5", Decimal("10"), now=T0)
    assert len(proc.ledger) == 5


def test_ledger_contains_txn_id(proc: CardProcessor) -> None:
    proc.process("card-1", "txn-abc", Decimal("10"), now=T0)
    assert proc.ledger[0]["txn_id"] == "txn-abc"


def test_ledger_contains_amount(proc: CardProcessor) -> None:
    proc.process("card-1", "txn-x", Decimal("123.45"), now=T0)
    assert proc.ledger[0]["amount"] == Decimal("123.45")


def test_exactly_four_transactions_not_rate_limited(proc: CardProcessor) -> None:
    for i in range(4):
        proc.process("card-1", f"txn-{i}", Decimal("10"), now=T0)
    assert proc.is_rate_limited(now=T0) is False


def test_rate_limit_boundary_exactly_five(proc: CardProcessor) -> None:
    for i in range(5):
        proc.process("card-1", f"txn-{i}", Decimal("10"), now=T0)
    assert proc.is_rate_limited(now=T0) is True
