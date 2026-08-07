"""Tests for Q15 — Transaction Processing Engine."""

from decimal import Decimal

import pytest

from q15_transaction_engine import DuplicateTxnError, TransactionEngine, TxnState


@pytest.fixture
def engine() -> TransactionEngine:
    e = TransactionEngine()
    e.add_account("acc-a", Decimal("1000"))
    e.add_account("acc-b", Decimal("500"))
    return e


def test_commit_credits_account(engine: TransactionEngine) -> None:
    engine.commit("t1", "acc-a", Decimal("100"))
    assert engine.balance("acc-a") == Decimal("1100")


def test_commit_debits_account(engine: TransactionEngine) -> None:
    engine.commit("t2", "acc-a", Decimal("-200"))
    assert engine.balance("acc-a") == Decimal("800")


def test_commit_returns_committed_txn(engine: TransactionEngine) -> None:
    txn = engine.commit("t3", "acc-a", Decimal("50"))
    assert txn.state == TxnState.COMMITTED


def test_commit_duplicate_id_raises(engine: TransactionEngine) -> None:
    engine.commit("t4", "acc-a", Decimal("10"))
    with pytest.raises(DuplicateTxnError):
        engine.commit("t4", "acc-a", Decimal("10"))


def test_commit_unknown_account_raises(engine: TransactionEngine) -> None:
    with pytest.raises(KeyError):
        engine.commit("t5", "ghost", Decimal("10"))


def test_commit_overdraft_raises(engine: TransactionEngine) -> None:
    with pytest.raises(ValueError):
        engine.commit("t6", "acc-a", Decimal("-9999"))


def test_rollback_reverses_credit(engine: TransactionEngine) -> None:
    engine.commit("t7", "acc-a", Decimal("100"))
    engine.rollback("t7")
    assert engine.balance("acc-a") == Decimal("1000")


def test_rollback_reverses_debit(engine: TransactionEngine) -> None:
    engine.commit("t8", "acc-a", Decimal("-200"))
    engine.rollback("t8")
    assert engine.balance("acc-a") == Decimal("1000")


def test_rollback_sets_state(engine: TransactionEngine) -> None:
    engine.commit("t9", "acc-a", Decimal("50"))
    txn = engine.rollback("t9")
    assert txn.state == TxnState.ROLLED_BACK


def test_rollback_twice_raises(engine: TransactionEngine) -> None:
    engine.commit("t10", "acc-a", Decimal("50"))
    engine.rollback("t10")
    with pytest.raises(ValueError):
        engine.rollback("t10")


def test_transfer_moves_funds(engine: TransactionEngine) -> None:
    engine.transfer("tr1", "acc-a", "acc-b", Decimal("300"))
    assert engine.balance("acc-a") == Decimal("700")
    assert engine.balance("acc-b") == Decimal("800")


def test_transfer_atomic_on_insufficient_funds(engine: TransactionEngine) -> None:
    with pytest.raises(ValueError):
        engine.transfer("tr2", "acc-a", "acc-b", Decimal("9999"))
    assert engine.balance("acc-a") == Decimal("1000")
    assert engine.balance("acc-b") == Decimal("500")


def test_balance_initial(engine: TransactionEngine) -> None:
    assert engine.balance("acc-a") == Decimal("1000")


def test_add_account_returns_account(engine: TransactionEngine) -> None:
    acc = engine.add_account("acc-c", Decimal("0"))
    assert acc.account_id == "acc-c"
