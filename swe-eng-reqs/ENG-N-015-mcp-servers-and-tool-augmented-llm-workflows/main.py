#!/usr/bin/env python3
"""ENG-N-015: MCP servers and tool-augmented LLM workflows

Python vertical-slice MVP for MCP/LLM/agent DevEx requirements.
Runs offline with deterministic fixtures (no paid LLM required).
"""

from __future__ import annotations

import json
import os
from http.server import BaseHTTPRequestHandler, HTTPServer
from pathlib import Path

from approval import grant
from mcp_mutating import MutatingMCPServer, handle_mcp

ROOT = Path(__file__).resolve().parent
STATE = {"runs": 0}
SERVER = MutatingMCPServer()


def info() -> dict:
    return {
        "requirement_id": "ENG-N-015",
        "service": "eng-n-015",
        "mcp_inspired": True,
        "mcp_sdk": False,
        "policy_gateway": True,
        "simulator": True,
        "note": "stdlib MCP-inspired subset; not the official MCP SDK",
    }


def demo_payload() -> dict:
    STATE["runs"] += 1
    server = MutatingMCPServer()
    arguments = {"pipeline": "build-42"}
    denied = server.call_tool("restart_pipeline", arguments)
    approval = grant(server.gateway.approvals, "restart_pipeline", arguments)
    allowed = server.call_tool(
        "restart_pipeline", arguments, approval["approval_token"]
    )
    return {
        "ok": True,
        **info(),
        "run": STATE["runs"],
        "mutate_denied_without_approval": denied.get("reason")
        == "approval_required",
        "mutate_allowed_with_approval": allowed.get("restarted") == "build-42",
        "audit_entries": len(server.gateway.audit_entries),
        "audit_tail": server.gateway.audit_entries[-3:],
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
            body = f"eng-n-015_demo_runs_total {STATE['runs']}\n".encode()
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
        try:
            length = int(self.headers.get("Content-Length", "0"))
            if length < 0 or length > 65_536:
                self._json(413, {"error": "request body too large"})
                return
            payload = json.loads(self.rfile.read(length) or b"{}")
            if not isinstance(payload, dict):
                raise ValueError("JSON body must be an object")
        except (ValueError, json.JSONDecodeError) as exc:
            self._json(400, {"error": str(exc)})
            return
        if self.path == "/v1/mcp":
            self._json(200, handle_mcp(payload, SERVER))
            return
        if self.path == "/v1/approvals":
            name = payload.get("name")
            args = payload.get("arguments", {})
            if not isinstance(name, str) or not isinstance(args, dict):
                self._json(400, {"error": "name and object arguments are required"})
                return
            try:
                tool = SERVER.gateway.get(name)
            except KeyError as exc:
                self._json(404, {"error": str(exc)})
                return
            if not tool.mutating:
                self._json(400, {"error": "approval only applies to mutating tools"})
                return
            self._json(201, grant(SERVER.gateway.approvals, name, args))
            return
        self._json(404, {"error": "not found"})

    def log_message(self, fmt: str, *args) -> None:
        return


def main() -> None:
    addr = os.environ.get("ADDR", "0.0.0.0")
    port = int(os.environ.get("PORT", "8080"))
    server = HTTPServer((addr, port), Handler)
    print(f"eng-n-015 listening on {addr}:{port} (requirement ENG-N-015)", flush=True)
    server.serve_forever()


if __name__ == "__main__":
    main()
