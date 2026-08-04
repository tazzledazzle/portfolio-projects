"""Tests for Q1 — Can Purchase Card."""

from decimal import Decimal

import pytest

from q01_can_purchase_card import Card


@pytest.fixture
def card() -> Card:
    return Card(card_id="c1", credit_limit=Decimal("1000"))


def test_can_purchase_within_limit(card: Card) -> None:
    assert card.can_purchase(Decimal("500")) is True


def test_can_purchase_exact_limit(card: Card) -> None:
    assert card.can_purchase(Decimal("1000")) is True


def test_can_purchase_exceeds_limit(card: Card) -> None:
    assert card.can_purchase(Decimal("1001")) is False


def test_available_credit_initial(card: Card) -> None:
    assert card.available_credit == Decimal("1000")


def test_process_purchase_reduces_balance(card: Card) -> None:
    card.process_purchase("txn-1", Decimal("300"))
    assert card.balance == Decimal("300")


def test_process_purchase_accepted_returns_true(card: Card) -> None:
    result = card.process_purchase("txn-2", Decimal("100"))
    assert result is True


def test_process_purchase_rejected_when_over_limit(card: Card) -> None:
    result = card.process_purchase("txn-3", Decimal("1500"))
    assert result is False


def test_available_credit_after_purchase(card: Card) -> None:
    card.process_purchase("txn-4", Decimal("400"))
    assert card.available_credit == Decimal("600")


def test_idempotent_duplicate_txn_id(card: Card) -> None:
    card.process_purchase("txn-5", Decimal("200"))
    result = card.process_purchase("txn-5", Decimal("200"))
    assert result is False


def test_idempotent_balance_not_doubled(card: Card) -> None:
    card.process_purchase("txn-6", Decimal("150"))
    card.process_purchase("txn-6", Decimal("150"))
    assert card.balance == Decimal("150")


def test_audit_log_populated_on_accept(card: Card) -> None:
    card.process_purchase("txn-7", Decimal("250"))
    assert len(card._audit_log) == 1


def test_audit_log_not_populated_on_reject(card: Card) -> None:
    card.process_purchase("txn-8", Decimal("9999"))
    assert len(card._audit_log) == 0


def test_reconcile_passes_after_purchases(card: Card) -> None:
    card.process_purchase("txn-9", Decimal("100"))
    card.process_purchase("txn-10", Decimal("200"))
    assert card.reconcile() is True


def test_reconcile_empty_card(card: Card) -> None:
    assert card.reconcile() is True


def test_multiple_purchases_accumulate_correctly(card: Card) -> None:
    card.process_purchase("txn-11", Decimal("300"))
    card.process_purchase("txn-12", Decimal("500"))
    assert card.balance == Decimal("800")
    assert card.available_credit == Decimal("200")
