#!/usr/bin/env python3
"""ENG-E-016: Agent frameworks familiarity

Python vertical-slice MVP for MCP/LLM/agent DevEx requirements.
Runs offline with deterministic fixtures (no paid LLM required).
"""

from __future__ import annotations

import json
import os
from http.server import BaseHTTPRequestHandler, HTTPServer

from agent_loop import FixturePlanner, run_agent
from tool_registry import Tool, ToolRegistry

STATE = {"runs": 0}


def info() -> dict:
    return {
        "requirement_id": "ENG-E-016",
        "service": "eng-e-016",
        "title": "Agent frameworks familiarity",
        "agent_framework_inspired": True,
        "simulator": True,
        "live_provider": False,
        "note": "Deterministic Plan-Execute simulator; no external agent framework or LLM.",
    }


def build_registry() -> ToolRegistry:
    registry = ToolRegistry()
    registry.register(
        Tool(
            name="inspect_pipeline",
            mutating=False,
            handler=lambda args: {
                "pipeline": args["pipeline"],
                "status": "failed",
                "finding": "test stage regression",
            },
            input_schema={"pipeline": "string"},
        )
    )
    return registry


def agent_payload(goal: str, max_steps: int = 5) -> dict:
    return run_agent(
        goal,
        build_registry(),
        FixturePlanner(),
        max_steps=max_steps,
    )


def demo_payload() -> dict:
    STATE["runs"] += 1
    result = agent_payload("explain pipeline build-42")
    return {
        "ok": True,
        "requirement_id": "ENG-E-016",
        "service": "eng-e-016",
        "title": "Agent frameworks familiarity",
        "run": STATE["runs"],
        "agent_framework_inspired": True,
        "simulator": True,
        "live_provider": False,
        **result,
        "acceptance": [
            "healthz returns ok",
            "fixture planner emits at least one step",
            "registered tools execute through the bounded loop",
            "trajectory passes deterministic evaluation",
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

    def _request_json(self) -> dict:
        length = int(self.headers.get("Content-Length", "0"))
        if length <= 0 or length > 65_536:
            raise ValueError("request body must be between 1 and 65536 bytes")
        payload = json.loads(self.rfile.read(length))
        if not isinstance(payload, dict):
            raise ValueError("request body must be a JSON object")
        return payload

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
            body = f"eng-e-016_demo_runs_total {STATE['runs']}\n".encode()
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
        if self.path == "/v1/agent/run":
            try:
                payload = self._request_json()
                goal = payload.get("goal")
                max_steps = payload.get("max_steps", 5)
                self._json(200, {"ok": True, **agent_payload(goal, max_steps)})
            except (json.JSONDecodeError, ValueError) as exc:
                self._json(400, {"ok": False, "error": str(exc)})
            return
        self._json(404, {"error": "not found"})

    def log_message(self, fmt: str, *args) -> None:
        return


def main() -> None:
    addr = os.environ.get("ADDR", "0.0.0.0")
    port = int(os.environ.get("PORT", "8080"))
    server = HTTPServer((addr, port), Handler)
    print(f"eng-e-016 listening on {addr}:{port} (requirement ENG-E-016)", flush=True)
    server.serve_forever()


if __name__ == "__main__":
    main()
