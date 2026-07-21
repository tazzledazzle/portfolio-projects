#!/usr/bin/env python3
"""ENG-N-014: Agentic systems with DevEx value

Python vertical-slice MVP for MCP/LLM/agent DevEx requirements.
Runs offline with deterministic fixtures (no paid LLM required).
"""

from __future__ import annotations

import json
import os
from http.server import BaseHTTPRequestHandler, HTTPServer
from pathlib import Path

from roi import compute_roi
from value_feature import run_value_feature

ROOT = Path(__file__).resolve().parent
STATE = {"runs": 0}


def info() -> dict:
    return {
        "requirement_id": "ENG-N-014",
        "service": "eng-n-014",
        "fixture_roi": True,
        "baseline_source": "fixture",
        "fabricated_prod": False,
        "live_provider": False,
    }


def demo_payload(_request: dict | None = None) -> dict:
    STATE["runs"] += 1
    fixture = json.loads((ROOT / "testdata" / "baselines.json").read_text())
    baseline = fixture["baseline"]
    assisted = run_value_feature(baseline)
    roi = compute_roi(baseline, assisted)
    return {
        "ok": True,
        **info(),
        **roi,
        "run": STATE["runs"],
        "assisted_path": assisted,
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
            body = f"eng-n-014_demo_runs_total {STATE['runs']}\n".encode()
            self.send_response(200)
            self.send_header("Content-Type", "text/plain")
            self.end_headers()
            self.wfile.write(body)
            return
        self._json(404, {"error": "not found"})

    def do_POST(self) -> None:  # noqa: N802
        if self.path == "/v1/demo":
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
            self._json(200, demo_payload(payload))
            return
        self._json(404, {"error": "not found"})

    def log_message(self, fmt: str, *args) -> None:
        return


def main() -> None:
    addr = os.environ.get("ADDR", "0.0.0.0")
    port = int(os.environ.get("PORT", "8080"))
    server = HTTPServer((addr, port), Handler)
    print(f"eng-n-014 listening on {addr}:{port} (requirement ENG-N-014)", flush=True)
    server.serve_forever()


if __name__ == "__main__":
    main()
