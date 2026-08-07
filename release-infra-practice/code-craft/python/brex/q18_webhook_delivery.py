"""
Q18 — Rate-limited Webhook Delivery with Retry/Backoff

Base: deliver a webhook event to a subscriber URL via an injected HTTP caller.
Ext 1: retry with exponential backoff on failure (max_retries attempts).
Ext 2: handle a subscriber that is intermittently down — track consecutive failures
       and mark subscriber as suspended after a threshold.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from typing import Protocol


class HttpPoster(Protocol):
    def post(self, url: str, payload: dict[str, object]) -> int:
        """Send POST; return HTTP status code."""
        ...


@dataclass
class DeliveryResult:
    success: bool
    attempts: int
    final_status: int | None = None
    error: str | None = None


@dataclass
class Subscriber:
    subscriber_id: str
    url: str
    consecutive_failures: int = 0
    suspended: bool = False
    SUSPEND_THRESHOLD: int = field(default=5, repr=False)


def deliver(
    event: dict[str, object],
    subscriber: Subscriber,
    http: HttpPoster,
    *,
    max_retries: int = 3,
    backoff_base: float = 0.0,  # 0 for tests to avoid sleeping
) -> DeliveryResult:
    """
    Attempt to deliver event to subscriber.url.

    - Retries up to max_retries times on non-2xx response.
    - On each consecutive failure, increments subscriber.consecutive_failures.
    - Suspends subscriber when consecutive_failures >= SUSPEND_THRESHOLD.
    - Resets consecutive_failures to 0 on success.
    - Returns immediately with error if subscriber is already suspended.
    """
    raise NotImplementedError
