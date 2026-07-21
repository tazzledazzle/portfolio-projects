"""Typed in-process tool registry with deny-by-default mutation approval."""

from __future__ import annotations

from dataclasses import dataclass
from typing import Any, Callable


@dataclass(frozen=True)
class Tool:
    name: str
    mutating: bool
    handler: Callable[[dict], Any]
    input_schema: dict


class ToolRegistry:
    def __init__(self) -> None:
        self._tools: dict[str, Tool] = {}

    def register(self, tool: Tool) -> None:
        if not tool.name.strip():
            raise ValueError("tool name must not be empty")
        if tool.name in self._tools:
            raise ValueError(f"tool already registered: {tool.name}")
        self._tools[tool.name] = tool

    def call(self, name: str, args: dict, *, approved: bool = False) -> dict:
        tool = self._tools.get(name)
        if tool is None:
            return {"isError": True, "reason": "unknown_tool"}
        if not isinstance(args, dict):
            return {"isError": True, "reason": "invalid_arguments"}
        missing = [field for field in tool.input_schema if field not in args]
        if missing:
            return {
                "isError": True,
                "reason": "invalid_arguments",
                "missing": missing,
            }
        if tool.mutating and not approved:
            return {"isError": True, "reason": "approval_required"}
        try:
            content = tool.handler(args)
        except (KeyError, TypeError, ValueError):
            return {"isError": True, "reason": "handler_rejected_arguments"}
        return {"isError": False, "content": content}
