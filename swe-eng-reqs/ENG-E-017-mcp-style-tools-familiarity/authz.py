"""Token-inspired authorization for read-only MCP-style tool calls."""

from __future__ import annotations

from dataclasses import dataclass


@dataclass(frozen=True)
class Authorization:
    allowed: bool
    subject: str | None = None
    reason: str | None = None


def authorize(token: str | None, required_scopes: set[str]) -> Authorization:
    """Authorize ``Bearer <subject>:<comma-separated scopes>`` demo tokens."""
    if not token or not token.startswith("Bearer "):
        return Authorization(False, reason="invalid_token")

    credential = token.removeprefix("Bearer ").strip()
    if ":" not in credential:
        return Authorization(False, reason="invalid_token")

    subject, scope_text = credential.split(":", 1)
    scopes = {scope.strip() for scope in scope_text.split(",") if scope.strip()}
    if not subject or not scopes:
        return Authorization(False, reason="invalid_token")
    if not required_scopes.issubset(scopes):
        return Authorization(False, subject=subject, reason="missing_scope")
    return Authorization(True, subject=subject)
