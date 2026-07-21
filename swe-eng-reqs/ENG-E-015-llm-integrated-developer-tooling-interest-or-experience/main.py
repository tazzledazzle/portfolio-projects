#!/usr/bin/env python3
"""ENG-E-015: LLM-integrated developer tooling interest or experience

Python vertical-slice MVP for MCP/LLM/agent DevEx requirements.
Runs offline with deterministic fixtures (no paid LLM required).
"""

from __future__ import annotations

import json
import os
from http.server import BaseHTTPRequestHandler, HTTPServer
from pathlib import Path

from summarizer import load_failure_fixtures, summarize_failures

ROOT = Path(__file__).resolve().parent
STATE = {"runs": 0, "audits": []}


def info() -> dict:
    return {
        "requirement_id": "ENG-E-015",
        "service": "eng-e-015",
        "title": "LLM-integrated developer tooling interest or experience",
        "offline_fixture_llm": True,
        "simulator": True,
        "live_provider": False,
        "note": "Deterministic offline fixture LLM; no live provider or API key.",
    }


def demo_payload() -> dict:
    STATE["runs"] += 1
    result = summarize_failures(
        load_failure_fixtures(ROOT / "testdata" / "failures.json")
    )
    STATE["audits"].append(
        {
            "action": "offline_summarize",
            "requirement_id": "ENG-E-015",
            "failures_ingested": result["failures_ingested"],
        }
    )
    return {
        "ok": True,
        "requirement_id": "ENG-E-015",
        "service": "eng-e-015",
        "title": "LLM-integrated developer tooling interest or experience",
        "run": STATE["runs"],
        "offline_fixture_llm": True,
        "offline_llm": result["offline_llm"],
        "fixture_mode": result["fixture_mode"],
        "live_provider": result["live_provider"],
        "simulator": True,
        "summary": result["summary"],
        "findings": result["findings"],
        "failures_ingested": result["failures_ingested"],
        "audit_tail": STATE["audits"][-3:],
        "acceptance": [
            "healthz returns ok",
            "fixture pipeline failures are summarized",
            "offline fixture LLM requires zero API keys",
            "live provider remains disabled",
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
            body = f"eng-e-015_demo_runs_total {STATE['runs']}\n".encode()
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
        if self.path == "/v1/summarize":
            try:
                payload = self._request_json()
                self._json(200, {"ok": True, **summarize_failures(payload)})
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
    print(f"eng-e-015 listening on {addr}:{port} (requirement ENG-E-015)", flush=True)
    server.serve_forever()


if __name__ == "__main__":
    main()
