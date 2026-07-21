"""Deny-by-default policy gateway for mutating MCP-inspired tools."""

from __future__ import annotations

from dataclasses import dataclass
from typing import Callable

from approval import ApprovalStore, intent_digest

SENSITIVE_KEYS = {"secret", "token", "password", "api_key"}


def _redact(value):
    if isinstance(value, dict):
        return {
            key: "[REDACTED]" if key.lower() in SENSITIVE_KEYS else _redact(item)
            for key, item in value.items()
        }
    if isinstance(value, list):
        return [_redact(item) for item in value]
    return value


@dataclass(frozen=True)
class Tool:
    name: str
    mutating: bool
    handler: Callable[[dict], dict]
    description: str
    input_schema: dict


class PolicyGateway:
    def __init__(self, approvals: ApprovalStore | None = None) -> None:
        self.approvals = approvals or ApprovalStore()
        self._tools: dict[str, Tool] = {}
        self.audit_entries: list[dict] = []

    def register(
        self,
        name: str,
        *,
        mutating: bool,
        handler: Callable[[dict], dict],
        description: str = "",
        input_schema: dict | None = None,
    ) -> None:
        self._tools[name] = Tool(
            name=name,
            mutating=mutating,
            handler=handler,
            description=description,
            input_schema=input_schema or {"type": "object"},
        )

    def get(self, name: str) -> Tool:
        if name not in self._tools:
            raise KeyError(f"unknown tool: {name}")
        return self._tools[name]

    def list_tools(self) -> list[dict]:
        return [
            {
                "name": tool.name,
                "description": tool.description,
                "inputSchema": tool.input_schema,
                "mutating": tool.mutating,
                "requires_approval": tool.mutating,
            }
            for tool in self._tools.values()
        ]

    def audit(self, decision: str, name: str, args: dict, digest: str) -> None:
        self.audit_entries.append(
            {
                "decision": decision,
                "tool": name,
                "intent_digest": digest,
                "args": _redact(args),
            }
        )


def call_with_policy(
    gateway: PolicyGateway,
    name: str,
    args: dict,
    approval_token: str | None = None,
) -> dict:
    tool = gateway.get(name)
    digest = intent_digest(name, args)
    if not tool.mutating:
        gateway.audit("allow_read", name, args, digest)
        return tool.handler(args)
    if not gateway.approvals.valid(approval_token, digest):
        gateway.audit("deny_mutate", name, args, digest)
        return {
            "isError": True,
            "reason": "approval_required",
            "intent_digest": digest,
        }
    gateway.audit("allow_mutate", name, args, digest)
    return tool.handler(args)
