"""
Q7 — Full-Stack Transactions App (backend logic layer)

The UI/full-stack build is out-of-scope for pure Python; this module captures the
core backend service logic: create accounts, post transactions, query history.
Base: create account, post a monetary transaction between two accounts.
Ext 1: query transaction history for an account.
Ext 2: enforce that the sender has sufficient funds; surface a structured error.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from datetime import datetime
from decimal import Decimal


@dataclass
class AppAccount:
    account_id: str
    owner: str
    balance: Decimal = Decimal("0")


@dataclass
class AppTransaction:
    txn_id: str
    sender_id: str
    receiver_id: str
    amount: Decimal
    created_at: datetime = field(default_factory=datetime.utcnow)


class InsufficientFundsError(Exception):
    pass


class TransactionApp:
    def __init__(self) -> None:
        self._accounts: dict[str, AppAccount] = {}
        self._transactions: list[AppTransaction] = []

    def create_account(self, account_id: str, owner: str, initial_balance: Decimal = Decimal("0")) -> AppAccount:
        """Create and register a new account. Raises ValueError if id already exists."""
        raise NotImplementedError

    def get_account(self, account_id: str) -> AppAccount:
        """Raise KeyError if not found."""
        raise NotImplementedError

    def post_transaction(self, txn_id: str, sender_id: str, receiver_id: str, amount: Decimal) -> AppTransaction:
        """
        Move amount from sender to receiver.

        Raises:
            InsufficientFundsError: if sender balance < amount.
            KeyError: if either account does not exist.
            ValueError: if amount <= 0.
        """
        raise NotImplementedError

    def history(self, account_id: str) -> list[AppTransaction]:
        """Return all transactions where account_id is sender or receiver, oldest first."""
        raise NotImplementedError
