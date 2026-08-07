"""
Q14 — P2P Payment App (backend service layer)

PayPal-like peer-to-peer payments: register users, send money, request money,
query history. This captures the business logic layer independent of UI.

Base: register users, send payment from sender to receiver.
Ext 1: request payments — a request can be accepted or declined.
Ext 2: full payment history per user with status tracking.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from datetime import datetime
from decimal import Decimal
from enum import Enum


class PaymentStatus(Enum):
    COMPLETED = "completed"
    PENDING = "pending"
    DECLINED = "declined"


@dataclass
class User:
    user_id: str
    name: str
    balance: Decimal = Decimal("0")


@dataclass
class Payment:
    payment_id: str
    sender_id: str
    receiver_id: str
    amount: Decimal
    status: PaymentStatus
    created_at: datetime = field(default_factory=datetime.utcnow)


class P2PService:
    def __init__(self) -> None:
        self._users: dict[str, User] = {}
        self._payments: list[Payment] = []

    def register(self, user_id: str, name: str, initial_balance: Decimal = Decimal("0")) -> User:
        """Raise ValueError if user_id already exists."""
        raise NotImplementedError

    def send(self, payment_id: str, sender_id: str, receiver_id: str, amount: Decimal) -> Payment:
        """
        Immediately transfer amount from sender to receiver.

        Raises:
            ValueError: amount <= 0 or sender == receiver.
            KeyError: unknown user.
            InsufficientFundsError: sender balance < amount.
        """
        raise NotImplementedError

    def request(self, payment_id: str, requester_id: str, payer_id: str, amount: Decimal) -> Payment:
        """Create a pending payment request. No funds moved yet."""
        raise NotImplementedError

    def accept_request(self, payment_id: str) -> Payment:
        """Complete a pending request, moving funds. Raises if not pending."""
        raise NotImplementedError

    def decline_request(self, payment_id: str) -> Payment:
        """Mark a pending request declined. No funds moved."""
        raise NotImplementedError

    def history(self, user_id: str) -> list[Payment]:
        """All payments involving user_id, oldest first."""
        raise NotImplementedError


class InsufficientFundsError(Exception):
    pass
