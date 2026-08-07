"""
Q10 — Idempotent Card-Transaction Processor

Base: accept a transaction; reject duplicates by idempotency key.
Ext 1: rate-limit — reject if the card has submitted >= 5 transactions in the last 60 s.
Ext 2: ledger — every accepted transaction is written to a simple append-only log.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from datetime import datetime, timedelta, timezone
from decimal import Decimal

RATE_LIMIT_COUNT: int = 5
RATE_LIMIT_WINDOW: timedelta = timedelta(seconds=60)


@dataclass
class CardProcessor:
    _seen_keys: set[str] = field(default_factory=set, repr=False)
    _timestamps: list[datetime] = field(default_factory=list, repr=False)
    _ledger: list[dict[str, object]] = field(default_factory=list, repr=False)

    def process(
        self,
        card_id: str,
        txn_id: str,
        amount: Decimal,
        *,
        now: datetime | None = None,
    ) -> str:
        """
        Attempt to process a card transaction.

        Returns one of: "accepted", "duplicate", "rate_limited".
        Accepted transactions are appended to the ledger.
        """
        raise NotImplementedError

    def is_rate_limited(self, *, now: datetime | None = None) -> bool:
        """Return True if the card has >= RATE_LIMIT_COUNT txns in the last window."""
        raise NotImplementedError

    @property
    def ledger(self) -> list[dict[str, object]]:
        return list(self._ledger)
