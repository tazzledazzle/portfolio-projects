"""
Q4 — Credit-limit Checker

Base: given an account's credit limit and outstanding balance, determine
      whether a new charge is approvable.
Ext 1: support multiple accounts; look up account by id.
Ext 2: support temporary credit-limit overrides (e.g. for high-value merchants)
       that expire after a single use.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from decimal import Decimal


@dataclass
class Account:
    account_id: str
    credit_limit: Decimal
    balance: Decimal = Decimal("0")
    _override_limit: Decimal | None = field(default=None, repr=False)

    def can_charge(self, amount: Decimal) -> bool:
        """Return True if amount fits within available credit (considering any override)."""
        if (self.available_credit - amount) >= 0:
            return True
        elif self._override_limit and self._override_limit + self.available_credit >= amount:
            self.set_override_limit(None)
            return True
        return False
    def charge(self, amount: Decimal) -> bool:
        """
        Attempt to apply a charge. Returns True on success, False if declined.
        A one-time override limit is consumed on the first successful charge that
        required it, then removed.
        """
        if self.can_charge(amount):
            self.balance += amount
            return True
        elif self._override_limit and self._override_limit + self.available_credit >= amount:
            self.set_override_limit(None)
            return True
        return False
    def set_override_limit(self, new_limit: Decimal) -> None:
        """Apply a temporary single-use credit-limit override."""
        self._override_limit = new_limit


    @property
    def available_credit(self) -> Decimal:

        if self._override_limit and self.balance > self.credit_limit:
            charge = self._override_limit
            self.set_override_limit(None)
            return (self.credit_limit + charge) - self.balance
        return self.credit_limit - self.balance


class CreditService:
    """Registry of accounts with bulk look-up support."""

    def __init__(self) -> None:
        self._accounts: dict[str, Account] = {}

    def add_account(self, account: Account) -> None:
        self._accounts[account.account_id] = account

    def get_account(self, account_id: str) -> Account:
        """Raise KeyError if account_id not found."""
        account = self._accounts.get(account_id, None)
        if not account:
            raise KeyError(f"Account {account_id} not found.")
        return account


    def can_charge(self, account_id: str, amount: Decimal) -> bool:
        # can find the account by id
        account = self.get_account(account_id)
        print(f'account: {account_id}, amount: {amount}')
        return account and account.can_charge(amount)

    def charge(self, account_id: str, amount: Decimal) -> bool:
        if self.can_charge(account_id, amount):
            self._accounts[account_id].balance += amount
            return True
        else:
            return False
