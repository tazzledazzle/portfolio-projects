"""
Category B -- parse + establish order pattern.
Run: python3 variations/kata_b_parse_sort.py
"""

from __future__ import annotations

import heapq
import re
from datetime import datetime, timedelta, timezone
from typing import Optional


def _parse_ts(raw: str) -> datetime:
    if raw.endswith("Z"):
        raw = raw[:-1] + "+00:00"
    dt = datetime.fromisoformat(raw)
    if dt.tzinfo is None:
        dt = dt.replace(tzinfo=timezone.utc)
    return dt.astimezone(timezone.utc)


# B1 ---------------------------------------------------------------------
def merge_paginated_commits(pages: list[list[dict]]) -> list[dict]:
    """
    Each page is already sorted by committed_at ascending; pages themselves
    may arrive in any order and overlap. k-way merge + dedup by sha.
    Time: O(n log k), Space: O(n + k).
    """
    streams = [
        iter(sorted(page, key=lambda c: _parse_ts(c["committed_at"])))
        for page in pages
    ]
    merged = heapq.merge(*streams, key=lambda c: _parse_ts(c["committed_at"]))
    seen: set[str] = set()
    result = []
    for commit in merged:
        if commit["sha"] not in seen:
            seen.add(commit["sha"])
            result.append(commit)
    return result


# B2 ---------------------------------------------------------------------
def normalize_mixed_format_transactions(raw: list[dict]) -> tuple[list[dict], list[dict]]:
    """Returns (valid_sorted, errors)."""
    valid, errors = [], []
    for tx in raw:
        ts_raw = tx["timestamp"]
        dt = None
        try:
            if isinstance(ts_raw, (int, float)):
                dt = datetime.fromtimestamp(ts_raw, tz=timezone.utc)
            elif isinstance(ts_raw, str) and ts_raw.isdigit():
                dt = datetime.fromtimestamp(int(ts_raw), tz=timezone.utc)
            elif isinstance(ts_raw, str) and re.match(r"^\d{2}/\d{2}/\d{4} \d{2}:\d{2}$", ts_raw):
                dt = datetime.strptime(ts_raw, "%m/%d/%Y %H:%M").replace(tzinfo=timezone.utc)
            else:
                dt = _parse_ts(ts_raw)
        except (ValueError, TypeError):
            errors.append(tx)
            continue
        valid.append({**tx, "_parsed_ts": dt})
    valid.sort(key=lambda t: t["_parsed_ts"])
    return valid, errors


# B3 ---------------------------------------------------------------------
def flatten_and_order_thread_replies(threads: list[dict]) -> list[dict]:
    """
    threads: [{"ts": str, "text": str, "replies": [{"ts": str, "text": str}, ...]}]
    Parent must sort before its own replies even on exact timestamp ties.
    """
    flat = []
    for order, thread in enumerate(threads):
        flat.append({**thread, "_ts": _parse_ts(thread["ts"]), "_depth": 0, "_thread": order})
        for reply in thread.get("replies", []):
            flat.append({**reply, "_ts": _parse_ts(reply["ts"]), "_depth": 1, "_thread": order})
    # sort by time, then thread grouping, then depth (parent=0 before reply=1)
    flat.sort(key=lambda m: (m["_ts"], m["_thread"], m["_depth"]))
    return flat


# B4 ---------------------------------------------------------------------
_SEVERITY_RANK = {"CRITICAL": 0, "ERROR": 1, "WARN": 2, "INFO": 3}
_LOG_RE = re.compile(r"^\[(?P<ts>[^\]]+)\]\s+(?P<level>\w+):\s+(?P<msg>.*)$")


def parse_and_rank_log_severity(lines: list[str]) -> list[dict]:
    parsed = []
    for line in lines:
        m = _LOG_RE.match(line)
        if not m:
            continue
        level = m.group("level")
        parsed.append({
            "level": level,
            "timestamp": _parse_ts(m.group("ts")),
            "message": m.group("msg"),
        })
    parsed.sort(key=lambda e: (_SEVERITY_RANK.get(e["level"], 99), e["timestamp"]))
    return parsed


# B5 ---------------------------------------------------------------------
def convert_and_sort_multi_currency_ledger(
    transactions: list[dict], rates: dict[str, float]
) -> tuple[list[dict], list[dict]]:
    """Returns (converted_sorted_desc, skipped_missing_rate)."""
    converted, skipped = [], []
    for tx in transactions:
        rate = rates.get(tx["currency"])
        if rate is None:
            skipped.append(tx)
            continue
        converted.append({**tx, "usd_amount": tx["amount"] * rate})
    converted.sort(key=lambda t: t["usd_amount"], reverse=True)
    return converted, skipped


# B6 ---------------------------------------------------------------------
def resolve_recurring_calendar_events(
    events: list[dict], range_start: datetime, range_end: datetime
) -> list[dict]:
    """
    events: [{"title": str, "start": datetime, "recurrence": {"freq": "daily"|"weekly", "count": int} | None}]
    """
    step_for = {"daily": timedelta(days=1), "weekly": timedelta(weeks=1)}
    occurrences = []
    for event in events:
        recurrence = event.get("recurrence")
        if not recurrence:
            if range_start <= event["start"] <= range_end:
                occurrences.append({"title": event["title"], "start": event["start"]})
            continue
        step = step_for[recurrence["freq"]]
        occ_time = event["start"]
        for _ in range(recurrence["count"]):
            if range_start <= occ_time <= range_end:
                occurrences.append({"title": event["title"], "start": occ_time})
            occ_time += step
    occurrences.sort(key=lambda o: o["start"])
    return occurrences


# B7 ---------------------------------------------------------------------
def dedupe_and_sort_api_pages_by_etag(records: list[dict]) -> list[dict]:
    """records: [{"id": str, "etag": str, "rank": int, ...}]"""
    latest_by_id: dict[str, dict] = {}
    for r in records:
        existing = latest_by_id.get(r["id"])
        if existing is None or r["etag"] > existing["etag"]:
            latest_by_id[r["id"]] = r
    result = list(latest_by_id.values())
    result.sort(key=lambda r: (r["rank"], r["id"]))
    return result


def _run_tests():
    pages = [
        [{"sha": "a", "committed_at": "2026-01-01T00:00:00Z"},
         {"sha": "c", "committed_at": "2026-01-01T00:10:00Z"}],
        [{"sha": "b", "committed_at": "2026-01-01T00:05:00Z"},
         {"sha": "c", "committed_at": "2026-01-01T00:10:00Z"}],  # overlap dup
    ]
    merged = merge_paginated_commits(pages)
    assert [c["sha"] for c in merged] == ["a", "b", "c"]
    print("B1 merge_paginated_commits: OK")

    raw = [
        {"id": 1, "timestamp": "2026-01-01T00:10:00Z"},
        {"id": 2, "timestamp": 1767225600},          # epoch int
        {"id": 3, "timestamp": "01/01/2026 00:05"},   # US format
        {"id": 4, "timestamp": "not-a-date"},         # error
    ]
    valid, errors = normalize_mixed_format_transactions(raw)
    assert len(errors) == 1 and errors[0]["id"] == 4
    assert [v["id"] for v in valid] == sorted([v["id"] for v in valid], key=lambda i: [t for t in valid if t["id"] == i][0]["_parsed_ts"])
    print("B2 normalize_mixed_format_transactions: OK")

    threads = [{
        "ts": "2026-01-01T00:00:00Z", "text": "parent",
        "replies": [{"ts": "2026-01-01T00:00:00Z", "text": "reply"}],
    }]
    flat = flatten_and_order_thread_replies(threads)
    assert [m["text"] for m in flat] == ["parent", "reply"]
    print("B3 flatten_and_order_thread_replies: OK")

    lines = [
        "[2026-01-01T00:00:03Z] ERROR: disk full",
        "[2026-01-01T00:00:01Z] CRITICAL: node down",
        "[2026-01-01T00:00:02Z] INFO: heartbeat",
    ]
    ranked = parse_and_rank_log_severity(lines)
    assert [e["level"] for e in ranked] == ["CRITICAL", "ERROR", "INFO"]
    print("B4 parse_and_rank_log_severity: OK")

    txs = [
        {"amount": 100, "currency": "EUR"},
        {"amount": 100, "currency": "USD"},
        {"amount": 50, "currency": "XYZ"},  # missing rate
    ]
    rates = {"EUR": 1.1, "USD": 1.0}
    conv, skipped = convert_and_sort_multi_currency_ledger(txs, rates)
    assert len(skipped) == 1 and skipped[0]["currency"] == "XYZ"
    assert conv[0]["currency"] == "EUR"  # 110 > 100
    print("B5 convert_and_sort_multi_currency_ledger: OK")

    events = [{
        "title": "standup", "start": datetime(2026, 1, 1, 9, tzinfo=timezone.utc),
        "recurrence": {"freq": "daily", "count": 3},
    }]
    occ = resolve_recurring_calendar_events(
        events,
        datetime(2026, 1, 1, tzinfo=timezone.utc),
        datetime(2026, 1, 5, tzinfo=timezone.utc),
    )
    assert len(occ) == 3
    print("B6 resolve_recurring_calendar_events: OK")

    records = [
        {"id": "x", "etag": "e1", "rank": 2},
        {"id": "x", "etag": "e2", "rank": 2},  # newer etag wins
        {"id": "y", "etag": "e1", "rank": 1},
    ]
    result = dedupe_and_sort_api_pages_by_etag(records)
    assert len(result) == 2
    assert result[0]["id"] == "y"
    x_record = [r for r in result if r["id"] == "x"][0]
    assert x_record["etag"] == "e2"
    print("B7 dedupe_and_sort_api_pages_by_etag: OK")

    print("\nAll category B tests passed.")


if __name__ == "__main__":
    _run_tests()
