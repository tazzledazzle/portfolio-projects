#!/usr/bin/env python3
"""ENG-E-017: MCP-style tools familiarity

Python vertical-slice MVP for MCP/LLM/agent DevEx requirements.
Runs offline with deterministic fixtures (no paid LLM required).
"""

from __future__ import annotations

import json
import os
from http.server import BaseHTTPRequestHandler, HTTPServer

from audit import AuditLog
from mcp_server import MCPServer, handle_mcp

STATE = {"runs": 0}
MAX_BODY_BYTES = 64 * 1024


def build_server() -> MCPServer:
    server = MCPServer(audit=AuditLog())
    server.register_read_tool(
        "list_pipelines",
        "List recent delivery pipelines",
        lambda arguments: {
            "pipelines": [
                {"name": "build-main", "status": "passed"},
                {"name": "release-api", "status": "running"},
            ]
        },
    )
    server.register_read_tool(
        "get_test_flakes",
        "Read deterministic test flake counts",
        lambda arguments: {
            "suite": arguments.get("suite", "all"),
            "flakes": 2,
            "window": "fixture-7d",
        },
        input_schema={
            "type": "object",
            "properties": {"suite": {"type": "string"}},
        },
    )
    return server


MCP_SERVER = build_server()


def info() -> dict:
    return {
        "requirement_id": "ENG-E-017",
        "service": "eng-e-017",
        "title": "MCP-style tools familiarity",
        "mcp_inspired": True,
        "mcp_sdk": False,
        "read_only": True,
        "simulator": True,
        "note": "In-process tools/list|tools/call subset; no official MCP SDK",
    }


def demo_payload() -> dict:
    STATE["runs"] += 1
    server = build_server()
    listed = handle_mcp(
        {"jsonrpc": "2.0", "id": "demo-list", "method": "tools/list"},
        server,
    )
    called = handle_mcp(
        {
            "jsonrpc": "2.0",
            "id": "demo-call",
            "method": "tools/call",
            "params": {
                "name": "get_test_flakes",
                "arguments": {"suite": "unit", "token": "must-not-leak"},
            },
        },
        server,
        token="Bearer demo-reader:tools:read",
    )
    tools = listed["result"]["tools"]
    call_result = called["result"]
    return {
        "ok": True,
        "requirement_id": "ENG-E-017",
        "service": "eng-e-017",
        "title": "MCP-style tools familiarity",
        "run": STATE["runs"],
        "mcp_inspired": True,
        "mcp_sdk": False,
        "simulator": True,
        "read_only": all(tool["read_only"] for tool in tools),
        "tools_listed": len(tools),
        "auth_ok": call_result["isError"] is False,
        "audit_entries": len(server.audit.entries),
        "tools": tools,
        "call_result": call_result,
        "audit_tail": list(server.audit.entries[-3:]),
        "acceptance": [
            "tools/list exposes read-only tools",
            "tools/call requires token scope",
            "every call is audited with secret redaction",
            "mcp_sdk=false is explicit",
        ],
    }


class Handler(BaseHTTPRequestHandler):
    def _json(self, code: int, payload: dict) -> None:
        body = json.dumps(payload, indent=2).encode()
        self.send_response(code)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def do_GET(self) -> None:  # noqa: N802
        if self.path in ("/healthz", "/readyz"):
            self._json(200, {"status": "ok"})
            return
        if self.path == "/v1/info":
            self._json(200, info())
            return
        if self.path == "/v1/demo":
            self._json(200, demo_payload())
            return
        if self.path == "/metrics":
            body = f"eng-e-017_demo_runs_total {STATE['runs']}\n".encode()
            self.send_response(200)
            self.send_header("Content-Type", "text/plain")
            self.end_headers()
            self.wfile.write(body)
            return
        self._json(404, {"error": "not found"})

    def do_POST(self) -> None:  # noqa: N802
        if self.path == "/v1/demo":
            self._json(200, demo_payload())
            return
        if self.path == "/v1/mcp":
            message = self._read_json()
            if message is None:
                return
            response = handle_mcp(
                message,
                MCP_SERVER,
                token=self.headers.get("Authorization"),
            )
            self._json(200, response)
            return
        self._json(404, {"error": "not found"})

    def _read_json(self) -> dict | None:
        try:
            length = int(self.headers.get("Content-Length", "0"))
            if length < 0 or length > MAX_BODY_BYTES:
                self._json(413, {"error": "request body too large"})
                return None
            value = json.loads(self.rfile.read(length) or b"{}")
        except (ValueError, json.JSONDecodeError):
            self._json(400, {"error": "invalid json"})
            return None
        if not isinstance(value, dict):
            self._json(400, {"error": "json object required"})
            return None
        return value

    def log_message(self, fmt: str, *args) -> None:
        return


def main() -> None:
    addr = os.environ.get("ADDR", "0.0.0.0")
    port = int(os.environ.get("PORT", "8080"))
    server = HTTPServer((addr, port), Handler)
    print(f"eng-e-017 listening on {addr}:{port} (requirement ENG-E-017)", flush=True)
    server.serve_forever()


if __name__ == "__main__":
    main()
