"""
Q11 — REST + JSON + Sort Service

Base: fetch JSON from multiple sources, normalize datetime fields (ISO-8601),
      merge and sort results by timestamp.
Ext 1: handle sources that are unavailable (timeout / error) gracefully.
Ext 2: de-duplicate records with the same id across sources.

The HTTP layer is injected so it can be mocked in tests.
"""

from __future__ import annotations

from dataclasses import dataclass
from datetime import datetime, timezone
from typing import Protocol


@dataclass
class Record:
    id: str
    timestamp: datetime
    payload: dict[str, object]


class HttpClient(Protocol):
    def get(self, url: str) -> dict[str, object] | list[object]:
        """Perform a GET and return parsed JSON. Raise on HTTP/network error."""
        ...


def normalize_timestamp(raw: str) -> datetime:
    """
    Parse an ISO-8601 timestamp string into a timezone-aware UTC datetime.

    Raises:
        ValueError: if raw cannot be parsed.
    """
    raise NotImplementedError


def fetch_and_normalize(client: HttpClient, url: str) -> list[Record]:
    """
    GET url, parse the JSON array of objects, normalize each 'timestamp' field.

    Returns an empty list on any HTTP error (graceful degradation).
    Raises:
        ValueError: if the response is not a list of objects.
    """
    raise NotImplementedError


def merge_and_sort(sources: list[list[Record]], *, deduplicate: bool = True) -> list[Record]:
    """
    Merge record lists, optionally de-duplicate by id (keep first occurrence),
    and sort oldest-first by timestamp.
    """
    raise NotImplementedError
