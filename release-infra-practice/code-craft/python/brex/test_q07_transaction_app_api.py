"""Tests for Q7 — Full-Stack Transactions App (backend logic)."""

from decimal import Decimal

import pytest

from q07_transaction_app_api import AppAccount, InsufficientFundsError, TransactionApp


@pytest.fixture
def app() -> TransactionApp:
    a = TransactionApp()
    a.create_account("alice", "Alice", Decimal("500"))
    a.create_account("bob", "Bob", Decimal("100"))
    return a


def test_create_account_returns_account(app: TransactionApp) -> None:
    acc = app.create_account("carol", "Carol", Decimal("200"))
    assert isinstance(acc, AppAccount)


def test_create_duplicate_account_raises(app: TransactionApp) -> None:
    with pytest.raises(ValueError):
        app.create_account("alice", "Alice Again")


def test_get_account(app: TransactionApp) -> None:
    acc = app.get_account("alice")
    assert acc.owner == "Alice"


def test_get_missing_account_raises(app: TransactionApp) -> None:
    with pytest.raises(KeyError):
        app.get_account("nobody")


def test_post_transaction_moves_funds(app: TransactionApp) -> None:
    app.post_transaction("t1", "alice", "bob", Decimal("100"))
    assert app.get_account("alice").balance == Decimal("400")
    assert app.get_account("bob").balance == Decimal("200")


def test_post_transaction_returns_transaction(app: TransactionApp) -> None:
    txn = app.post_transaction("t2", "alice", "bob", Decimal("50"))
    assert txn.txn_id == "t2"


def test_insufficient_funds_raises(app: TransactionApp) -> None:
    with pytest.raises(InsufficientFundsError):
        app.post_transaction("t3", "bob", "alice", Decimal("999"))


def test_zero_amount_raises(app: TransactionApp) -> None:
    with pytest.raises(ValueError):
        app.post_transaction("t4", "alice", "bob", Decimal("0"))


def test_negative_amount_raises(app: TransactionApp) -> None:
    with pytest.raises(ValueError):
        app.post_transaction("t5", "alice", "bob", Decimal("-10"))


def test_unknown_sender_raises(app: TransactionApp) -> None:
    with pytest.raises(KeyError):
        app.post_transaction("t6", "ghost", "bob", Decimal("10"))


def test_history_contains_sent_txn(app: TransactionApp) -> None:
    app.post_transaction("t7", "alice", "bob", Decimal("20"))
    h = app.history("alice")
    assert any(t.txn_id == "t7" for t in h)


def test_history_contains_received_txn(app: TransactionApp) -> None:
    app.post_transaction("t8", "alice", "bob", Decimal("20"))
    h = app.history("bob")
    assert any(t.txn_id == "t8" for t in h)


def test_history_ordered_oldest_first(app: TransactionApp) -> None:
    app.post_transaction("t9", "alice", "bob", Decimal("10"))
    app.post_transaction("t10", "alice", "bob", Decimal("10"))
    h = app.history("alice")
    assert h[0].txn_id == "t9"


def test_history_empty_for_new_account(app: TransactionApp) -> None:
    app.create_account("dave", "Dave", Decimal("100"))
    assert app.history("dave") == []
