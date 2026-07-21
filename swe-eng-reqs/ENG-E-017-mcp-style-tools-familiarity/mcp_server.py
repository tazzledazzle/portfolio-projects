"""Read-only, MCP-inspired tools/list and tools/call simulator."""

from __future__ import annotations

from dataclasses import dataclass
from typing import Any, Callable

from audit import AuditLog
from authz import authorize

ToolHandler = Callable[[dict[str, Any]], dict[str, Any]]


@dataclass(frozen=True)
class Tool:
    name: str
    description: str
    handler: ToolHandler
    input_schema: dict[str, Any]
    mutating: bool = False

    def public_schema(self) -> dict[str, Any]:
        return {
            "name": self.name,
            "description": self.description,
            "inputSchema": self.input_schema,
            "mutating": self.mutating,
            "read_only": not self.mutating,
        }


class MCPServer:
    """Narrow protocol simulator that rejects mutating tool registration."""

    def __init__(self, *, audit: AuditLog | None = None) -> None:
        self.audit = audit or AuditLog()
        self._tools: dict[str, Tool] = {}

    def register_read_tool(
        self,
        name: str,
        description: str,
        handler: ToolHandler,
        input_schema: dict[str, Any] | None = None,
    ) -> None:
        self.register_tool(
            name,
            description,
            handler,
            input_schema=input_schema,
            mutating=False,
        )

    def register_tool(
        self,
        name: str,
        description: str,
        handler: ToolHandler,
        *,
        input_schema: dict[str, Any] | None = None,
        mutating: bool = False,
    ) -> None:
        if mutating:
            raise ValueError("ENG-E-017 is read-only; mutating tools are rejected")
        if not name or name in self._tools:
            raise ValueError("tool name must be non-empty and unique")
        self._tools[name] = Tool(
            name=name,
            description=description,
            handler=handler,
            input_schema=input_schema or {"type": "object", "properties": {}},
        )

    def list_tools(self) -> list[dict[str, Any]]:
        return [tool.public_schema() for tool in self._tools.values()]

    def call_tool(
        self,
        name: str,
        arguments: dict[str, Any],
        *,
        token: str | None,
    ) -> dict[str, Any]:
        authorization = authorize(token, {"tools:read"})
        if not authorization.allowed:
            self.audit.append(
                tool=name,
                decision="deny",
                arguments=arguments,
                subject=authorization.subject,
                reason=authorization.reason,
            )
            return {"isError": True, "reason": "unauthorized"}

        tool = self._tools.get(name)
        if tool is None:
            self.audit.append(
                tool=name,
                decision="deny",
                arguments=arguments,
                subject=authorization.subject,
                reason="unknown_tool",
            )
            return {"isError": True, "reason": "unknown_tool"}

        try:
            content = tool.handler(arguments)
        except (KeyError, TypeError, ValueError) as error:
            self.audit.append(
                tool=name,
                decision="deny",
                arguments=arguments,
                subject=authorization.subject,
                reason="invalid_arguments",
            )
            return {
                "isError": True,
                "reason": "invalid_arguments",
                "message": str(error),
            }

        self.audit.append(
            tool=name,
            decision="allow",
            arguments=arguments,
            subject=authorization.subject,
        )
        return {"isError": False, "content": content}


def handle_mcp(
    message: dict[str, Any],
    server: MCPServer,
    *,
    token: str | None = None,
) -> dict[str, Any]:
    request_id = message.get("id")
    method = message.get("method")
    if method == "tools/list":
        return {
            "jsonrpc": "2.0",
            "id": request_id,
            "result": {"tools": server.list_tools()},
        }
    if method == "tools/call":
        params = message.get("params")
        if not isinstance(params, dict) or not isinstance(params.get("name"), str):
            return _error(request_id, -32602, "invalid params")
        arguments = params.get("arguments", {})
        if not isinstance(arguments, dict):
            return _error(request_id, -32602, "invalid params")
        return {
            "jsonrpc": "2.0",
            "id": request_id,
            "result": server.call_tool(params["name"], arguments, token=token),
        }
    return _error(request_id, -32601, "method not found")


def _error(request_id: Any, code: int, message: str) -> dict[str, Any]:
    return {
        "jsonrpc": "2.0",
        "id": request_id,
        "error": {"code": code, "message": message},
    }
