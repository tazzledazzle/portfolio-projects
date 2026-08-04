"""
Q1 — Can Purchase Card

Base: decide whether a card purchase is allowed given a card's credit limit,
      current balance, and the requested amount.
      - allowed purchase = (credit line, current balance, request amount)
Ext 1: idempotency — re-submitting the same transaction_id must not double-charge.
Ext 2: audit log — every accepted transaction is appended to an audit ledger and
       can be reconciled (sum of ledger entries == current balance).
"""

from __future__ import annotations

from dataclasses import dataclass, field
from decimal import Decimal


@dataclass
class Card:
    card_id: str
    credit_limit: Decimal
    balance: Decimal = Decimal("0")
    _seen_txn_ids: set[str] = field(default_factory=set, repr=False)
    _audit_log: list[dict[str, object]] = field(default_factory=list, repr=False)

    def can_purchase(self, amount: Decimal) -> bool:
        """Return True if amount does not exceed available credit."""
        return (self.available_credit - amount) >= 0

    def process_purchase(self, txn_id: str, amount: Decimal) -> bool:
        """
        Attempt to process a purchase.

        Returns True if accepted, False if rejected or duplicate.
        Idempotent: a repeated txn_id is silently ignored (returns False).
        Accepted transactions are appended to the audit log.
        """
        if txn_id in self._seen_txn_ids:
            # print(f'DUPLICATE:::::{txn_id}': amount)
            return False
        if self.can_purchase(amount):
            self._seen_txn_ids.add(txn_id)
            self._audit_log.append({f'{txn_id}': amount})
            self.balance += amount
            return True
        return False

    def reconcile(self) -> bool:
        """Return True if sum of audit log amounts equals current balance."""
        return sum(list([sum(x.values()) for x  in  self._audit_log])) == self.balance

    @property
    def available_credit(self) -> Decimal:
        """Returns the credit limit sans current balance."""
        return self.credit_limit - self.balance

