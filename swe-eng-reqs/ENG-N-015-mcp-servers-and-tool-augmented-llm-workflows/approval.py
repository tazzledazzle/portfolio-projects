"""Digest-bound approval grants for mutating tool calls."""

from __future__ import annotations

import hashlib
import json
import secrets


def intent_digest(name: str, args: dict) -> str:
    """Return a stable digest for one exact tool-call intent."""
    canonical = json.dumps(
        {"name": name, "arguments": args},
        sort_keys=True,
        separators=(",", ":"),
        ensure_ascii=True,
    ).encode("utf-8")
    return hashlib.sha256(canonical).hexdigest()


class ApprovalStore:
    """In-memory store of opaque tokens bound to intent digests."""

    def __init__(self) -> None:
        self._grants: dict[str, str] = {}

    def grant(self, digest: str) -> str:
        token = f"appr_{secrets.token_hex(16)}"
        self._grants[token] = digest
        return token

    def valid(self, token: str | None, digest: str) -> bool:
        return bool(token) and self._grants.get(token) == digest


def grant(store: ApprovalStore, name: str, args: dict) -> dict:
    """Create a grant for an exact name/arguments pair."""
    digest = intent_digest(name, args)
    return {"approval_token": store.grant(digest), "intent_digest": digest}
