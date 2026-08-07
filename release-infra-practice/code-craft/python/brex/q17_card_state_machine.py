"""
Q17 — Card Freeze/Unfreeze State Machine

Base: a card has states ACTIVE | FROZEN; implement freeze/unfreeze transitions.
Ext 1: full audit trail of every state transition with timestamp and actor.
Ext 2: permanently closed state — a closed card cannot be frozen or unfrozen.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum


class CardState(Enum):
    ACTIVE = "active"
    FROZEN = "frozen"
    CLOSED = "closed"


@dataclass
class AuditEntry:
    from_state: CardState
    to_state: CardState
    actor: str
    timestamp: datetime = field(default_factory=datetime.utcnow)


@dataclass
class ManagedCard:
    card_id: str
    state: CardState = CardState.ACTIVE
    _audit: list[AuditEntry] = field(default_factory=list, repr=False)

    def freeze(self, actor: str) -> None:
        """
        Transition ACTIVE → FROZEN.
        Raise ValueError if already FROZEN or CLOSED.
        """
        raise NotImplementedError

    def unfreeze(self, actor: str) -> None:
        """
        Transition FROZEN → ACTIVE.
        Raise ValueError if not FROZEN or if CLOSED.
        """
        raise NotImplementedError

    def close(self, actor: str) -> None:
        """
        Transition any non-CLOSED state → CLOSED.
        Raise ValueError if already CLOSED.
        """
        raise NotImplementedError

    @property
    def audit_trail(self) -> list[AuditEntry]:
        """Return a copy of the audit trail."""
        raise NotImplementedError
