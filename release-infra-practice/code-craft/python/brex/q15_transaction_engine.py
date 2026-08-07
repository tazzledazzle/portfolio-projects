"""
Q15 — Transaction Processing Engine

Base: accept and commit transactions with exactly-once semantics.
Ext 1: rollback — a transaction can be reversed; balances are restored.
Ext 2: atomic multi-leg transfers (debit one account, credit another atomically).
"""

from __future__ import annotations

from dataclasses import dataclass, field
from decimal import Decimal
from enum import Enum


class TxnState(Enum):
    COMMITTED = "committed"
    ROLLED_BACK = "rolled_back"


@dataclass
class EngineAccount:
    account_id: str
    balance: Decimal = Decimal("0")


@dataclass
class Txn:
    txn_id: str
    account_id: str
    delta: Decimal  # positive = credit, negative = debit
    state: TxnState = TxnState.COMMITTED


class DuplicateTxnError(Exception):
    pass


class TransactionEngine:
    def __init__(self) -> None:
        self._accounts: dict[str, EngineAccount] = {}
        self._txns: dict[str, Txn] = {}

    def add_account(self, account_id: str, initial_balance: Decimal = Decimal("0")) -> EngineAccount:
        raise NotImplementedError

    def commit(self, txn_id: str, account_id: str, delta: Decimal) -> Txn:
        """
        Apply delta to account balance and record the transaction.

        Raises:
            DuplicateTxnError: if txn_id already exists.
            KeyError: if account_id not found.
            ValueError: if resulting balance would go negative.
        """
        raise NotImplementedError

    def rollback(self, txn_id: str) -> Txn:
        """
        Reverse a committed transaction. Raises ValueError if already rolled back.
        """
        raise NotImplementedError

    def transfer(self, txn_id: str, from_id: str, to_id: str, amount: Decimal) -> tuple[Txn, Txn]:
        """
        Atomically debit from_id and credit to_id.
        Either both succeed or neither is applied.
        """
        raise NotImplementedError

    def balance(self, account_id: str) -> Decimal:
        raise NotImplementedError
