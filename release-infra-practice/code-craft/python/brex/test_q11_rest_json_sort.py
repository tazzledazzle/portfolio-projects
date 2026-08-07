"""Tests for Q11 — REST + JSON + Sort Service."""

from datetime import datetime, timezone
from unittest.mock import MagicMock

import pytest

from q11_rest_json_sort import Record, fetch_and_normalize, merge_and_sort, normalize_timestamp


def test_normalize_iso8601_utc() -> None:
    dt = normalize_timestamp("2024-01-15T10:30:00Z")
    assert dt.tzinfo is not None
    assert dt.year == 2024


def test_normalize_iso8601_offset() -> None:
    dt = normalize_timestamp("2024-01-15T12:30:00+02:00")
    assert dt.tzinfo is not None


def test_normalize_converts_to_utc() -> None:
    dt = normalize_timestamp("2024-01-15T12:00:00+02:00")
    assert dt.hour == 10


def test_normalize_invalid_raises() -> None:
    with pytest.raises(ValueError):
        normalize_timestamp("NOT_A_DATE")


def make_client(data: list[dict]) -> MagicMock:
    client = MagicMock()
    client.get.return_value = data
    return client


def test_fetch_and_normalize_returns_records() -> None:
    client = make_client([
        {"id": "r1", "timestamp": "2024-01-01T00:00:00Z", "value": 1},
    ])
    records = fetch_and_normalize(client, "http://example.com")
    assert len(records) == 1
    assert isinstance(records[0], Record)


def test_fetch_normalizes_timestamp() -> None:
    client = make_client([
        {"id": "r1", "timestamp": "2024-06-15T10:00:00Z", "value": 1},
    ])
    records = fetch_and_normalize(client, "http://example.com")
    assert records[0].timestamp.tzinfo is not None


def test_fetch_returns_empty_on_error() -> None:
    client = MagicMock()
    client.get.side_effect = ConnectionError("unreachable")
    records = fetch_and_normalize(client, "http://down.example.com")
    assert records == []


def test_merge_and_sort_orders_by_timestamp() -> None:
    r1 = Record("a", datetime(2024, 1, 2, tzinfo=timezone.utc), {})
    r2 = Record("b", datetime(2024, 1, 1, tzinfo=timezone.utc), {})
    result = merge_and_sort([[r1, r2]])
    assert result[0].id == "b"


def test_merge_deduplicates_by_id() -> None:
    r1 = Record("dup", datetime(2024, 1, 1, tzinfo=timezone.utc), {})
    r2 = Record("dup", datetime(2024, 1, 2, tzinfo=timezone.utc), {})
    result = merge_and_sort([[r1], [r2]], deduplicate=True)
    assert len(result) == 1


def test_merge_no_dedup_keeps_all() -> None:
    r1 = Record("dup", datetime(2024, 1, 1, tzinfo=timezone.utc), {})
    r2 = Record("dup", datetime(2024, 1, 2, tzinfo=timezone.utc), {})
    result = merge_and_sort([[r1], [r2]], deduplicate=False)
    assert len(result) == 2


def test_merge_multiple_sources() -> None:
    r1 = Record("a", datetime(2024, 1, 1, tzinfo=timezone.utc), {})
    r2 = Record("b", datetime(2024, 1, 3, tzinfo=timezone.utc), {})
    r3 = Record("c", datetime(2024, 1, 2, tzinfo=timezone.utc), {})
    result = merge_and_sort([[r1, r2], [r3]])
    assert [r.id for r in result] == ["a", "c", "b"]


def test_merge_empty_sources() -> None:
    result = merge_and_sort([[], []])
    assert result == []


def test_fetch_passes_id_through() -> None:
    client = make_client([{"id": "xyz-42", "timestamp": "2024-01-01T00:00:00Z"}])
    records = fetch_and_normalize(client, "http://example.com")
    assert records[0].id == "xyz-42"
