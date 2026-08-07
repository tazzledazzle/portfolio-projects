"""
Q6 — Points-from-Transactions Service

1 USD = 1.5 loyalty points (use Decimal arithmetic; no float drift).
Base: award points for a USD transaction amount, query balance.
Ext 1: redeem points; prevent double-spend (balance may not go negative).
Ext 2: full transaction ledger with idempotency key per award/redemption.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from decimal import Decimal

POINTS_PER_USD: Decimal = Decimal("1.5")


@dataclass
class PointsAccount:
    account_id: str
    _balance: Decimal = field(default=Decimal("0"), repr=False)
    _ledger: list[dict[str, object]] = field(default_factory=list, repr=False)
    _seen_keys: set[str] = field(default_factory=set, repr=False)

    @property
    def balance(self) -> Decimal:
        raise NotImplementedError

    def award(self, usd_amount: Decimal, idempotency_key: str) -> Decimal:
        """
        Convert usd_amount to points and credit the account.
        Idempotent: repeated key returns 0 without re-crediting.
        Returns points awarded.
        """
        raise NotImplementedError

    def redeem(self, points: Decimal, idempotency_key: str) -> bool:
        """
        Debit points from balance.
        Returns False (no-op) if balance < points or key is duplicate.
        """
        raise NotImplementedError

    def ledger_sum(self) -> Decimal:
        """Return net sum of all ledger entries (should equal balance)."""
        raise NotImplementedError
