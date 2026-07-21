"""Append-only audit records with recursive secret redaction."""

from __future__ import annotations

import time
from copy import deepcopy
from typing import Any

SECRET_KEYS = ("api_key", "password", "secret", "token")


def redact(value: Any) -> Any:
    if isinstance(value, dict):
        return {
            key: (
                "[REDACTED]"
                if any(secret in key.lower() for secret in SECRET_KEYS)
                else redact(item)
            )
            for key, item in value.items()
        }
    if isinstance(value, list):
        return [redact(item) for item in value]
    return deepcopy(value)


class AuditLog:
    def __init__(self) -> None:
        self._entries: list[dict[str, Any]] = []

    @property
    def entries(self) -> tuple[dict[str, Any], ...]:
        return tuple(deepcopy(self._entries))

    def append(
        self,
        *,
        tool: str,
        decision: str,
        arguments: dict[str, Any],
        subject: str | None = None,
        reason: str | None = None,
    ) -> dict[str, Any]:
        entry = {
            "ts": time.time(),
            "tool": tool,
            "decision": decision,
            "subject": subject,
            "reason": reason,
            "arguments": redact(arguments),
        }
        self._entries.append(entry)
        return deepcopy(entry)
