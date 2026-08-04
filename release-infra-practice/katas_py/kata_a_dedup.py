"""
Category A -- dedup / first-repeat pattern.
Run: python3 variations/kata_a_dedup.py
"""

from __future__ import annotations

from collections import deque
from datetime import datetime, timezone
from typing import Optional


def _parse_ts(raw: str) -> datetime:
    if raw.endswith("Z"):
        raw = raw[:-1] + "+00:00"
    dt = datetime.fromisoformat(raw)
    if dt.tzinfo is None:
        dt = dt.replace(tzinfo=timezone.utc)
    return dt.astimezone(timezone.utc)


# A1 ---------------------------------------------------------------------
def first_replayed_request(log: list[dict]) -> Optional[str]:
    seen: set[str] = set()
    for entry in log:
        rid = entry["request_id"]
        if rid in seen:
            return rid
        seen.add(rid)
    return None


# A2 ---------------------------------------------------------------------
def first_ip_over_threshold(events: list[dict], k: int, window_seconds: int = 60) -> Optional[str]:
    """
    Sliding window per-IP. Time: O(n) amortized (each event enters/leaves
    each IP's deque at most once). Space: O(n) worst case across all IPs.
    """
    windows: dict[str, deque] = {}
    for event in events:
        ip = event["ip"]
        ts = _parse_ts(event["ts"])
        dq = windows.setdefault(ip, deque())
        dq.append(ts)
        while dq and (ts - dq[0]).total_seconds() > window_seconds:
            dq.popleft()
        if len(dq) > k:
            return ip
    return None


# A3 ---------------------------------------------------------------------
def first_duplicate_idempotency_key(messages: list[dict]) -> Optional[dict]:
    seen: dict[str, int] = {}
    for msg in messages:
        key = msg["idempotency_key"]
        amount = msg["amount_cents"]
        if key in seen:
            status = "duplicate" if seen[key] == amount else "conflict"
            return {"idempotency_key": key, "status": status}
        seen[key] = amount
    return None


# A4 ---------------------------------------------------------------------
def first_reused_session_token(events: list[dict]) -> Optional[str]:
    first_region: dict[str, str] = {}
    for event in events:
        token, region = event["token"], event["region"]
        if token in first_region:
            if first_region[token] != region:
                return token
        else:
            first_region[token] = region
    return None


# A5 ---------------------------------------------------------------------
def first_repeated_device_fingerprint(signups: list[dict]) -> Optional[str]:
    fp_emails: dict[str, set[str]] = {}
    for signup in signups:
        fp, email = signup["device_fingerprint"], signup["email"]
        emails = fp_emails.setdefault(fp, set())
        if emails and email not in emails:
            return fp
        emails.add(email)
    return None


# A6 ---------------------------------------------------------------------
def first_cycle_in_redirect_chain(urls: list[str]) -> Optional[str]:
    seen: set[str] = set()
    for url in urls:
        if url in seen:
            return url
        seen.add(url)
    return None


# A7 ---------------------------------------------------------------------
def first_stale_cache_key_collision(writes: list[dict]) -> Optional[dict]:
    max_version: dict[str, int] = {}
    for write in writes:
        key, version = write["key"], write["version"]
        if key in max_version and version < max_version[key]:
            return write
        max_version[key] = max(max_version.get(key, version), version)
    return None


def _run_tests():
    log = [{"request_id": "r1"}, {"request_id": "r2"}, {"request_id": "r1"}]
    assert first_replayed_request(log) == "r1"
    assert first_replayed_request([{"request_id": "x"}]) is None
    print("A1 first_replayed_request: OK")

    base = datetime(2026, 1, 1, tzinfo=timezone.utc)
    events = [
        {"ip": "1.1.1.1", "ts": base.isoformat()},
        {"ip": "1.1.1.1", "ts": base.replace(second=10).isoformat()},
        {"ip": "1.1.1.1", "ts": base.replace(second=20).isoformat()},
        {"ip": "1.1.1.1", "ts": base.replace(second=30).isoformat()},
    ]
    assert first_ip_over_threshold(events, k=2, window_seconds=60) == "1.1.1.1"
    assert first_ip_over_threshold(events, k=10, window_seconds=60) is None
    print("A2 first_ip_over_threshold: OK")

    msgs = [
        {"idempotency_key": "k1", "amount_cents": 500},
        {"idempotency_key": "k2", "amount_cents": 300},
        {"idempotency_key": "k1", "amount_cents": 500},
    ]
    result = first_duplicate_idempotency_key(msgs)
    assert result == {"idempotency_key": "k1", "status": "duplicate"}
    msgs_conflict = [
        {"idempotency_key": "k1", "amount_cents": 500},
        {"idempotency_key": "k1", "amount_cents": 999},
    ]
    assert first_duplicate_idempotency_key(msgs_conflict)["status"] == "conflict"
    print("A3 first_duplicate_idempotency_key: OK")

    sess = [
        {"token": "t1", "region": "us-east"},
        {"token": "t2", "region": "eu-west"},
        {"token": "t1", "region": "us-east"},
        {"token": "t1", "region": "ap-south"},
    ]
    assert first_reused_session_token(sess) == "t1"
    print("A4 first_reused_session_token: OK")

    signups = [
        {"device_fingerprint": "f1", "email": "a@x.com"},
        {"device_fingerprint": "f1", "email": "a@x.com"},  # same email, ok
        {"device_fingerprint": "f1", "email": "b@x.com"},  # flag
    ]
    assert first_repeated_device_fingerprint(signups) == "f1"
    print("A5 first_repeated_device_fingerprint: OK")

    chain = ["u1", "u2", "u3", "u2"]
    assert first_cycle_in_redirect_chain(chain) == "u2"
    print("A6 first_cycle_in_redirect_chain: OK")

    writes = [
        {"key": "k", "version": 3},
        {"key": "k", "version": 5},
        {"key": "k", "version": 4},  # stale
    ]
    stale = first_stale_cache_key_collision(writes)
    assert stale == {"key": "k", "version": 4}
    print("A7 first_stale_cache_key_collision: OK")

    print("\nAll category A tests passed.")


if __name__ == "__main__":
    _run_tests()
