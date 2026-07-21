#!/usr/bin/env python3
"""ENG-H-005: Eval/ops discipline for LLM/agent features

Python vertical-slice MVP for MCP/LLM/agent DevEx requirements.
Runs offline with deterministic fixtures (no paid LLM required).
"""

from __future__ import annotations

import json
import os
from http.server import BaseHTTPRequestHandler, HTTPServer
from pathlib import Path

from eval_harness import EvalHarness

ROOT = Path(__file__).resolve().parent
TESTDATA = ROOT / "testdata"
STATE = {"runs": 0}
MAX_BODY_BYTES = 64 * 1024


def info() -> dict:
    return {
        "requirement_id": "ENG-H-005",
        "service": "eng-h-005",
        "title": "Eval/ops discipline for LLM/agent features",
        "eval_ops": True,
        "simulator": True,
        "live_provider": False,
        "modes": ["offline", "online-sim"],
        "note": "Online-sim injects a local deterministic runner; no network LLM",
    }


def demo_payload() -> dict:
    STATE["runs"] += 1
    harness = EvalHarness(TESTDATA)
    offline = harness.run_offline()
    online_sim = harness.run_online_sim(
        lambda case: str(case["candidate"])
    )
    return {
        "ok": True,
        "requirement_id": "ENG-H-005",
        "service": "eng-h-005",
        "title": "Eval/ops discipline for LLM/agent features",
        "run": STATE["runs"],
        "eval_ops": True,
        "simulator": True,
        "live_provider": False,
        "mode": ["offline", "online-sim"],
        "offline_pass": offline["offline_pass"],
        "online_sim_pass": online_sim["online_sim_pass"],
        "failure_fixtures_caught": offline["failure_fixtures_caught"],
        "cases_run": offline["cases_run"],
        "reports": offline["reports"],
        "acceptance": [
            "golden fixtures pass offline",
            "known-bad fixtures fail and are caught",
            "online-sim uses only a local injected runner",
            "no live model provider or API key is required",
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
            body = f"eng-h-005_demo_runs_total {STATE['runs']}\n".encode()
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
        if self.path == "/v1/evals/run":
            request = self._read_json()
            if request is None:
                return
            harness = EvalHarness(TESTDATA)
            mode = request.get("mode", "offline")
            if mode == "offline":
                self._json(200, harness.run_offline())
                return
            if mode == "online-sim":
                self._json(
                    200,
                    harness.run_online_sim(
                        lambda case: str(case["candidate"])
                    ),
                )
                return
            self._json(400, {"error": "mode must be offline or online-sim"})
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
    print(f"eng-h-005 listening on {addr}:{port} (requirement ENG-H-005)", flush=True)
    server.serve_forever()


if __name__ == "__main__":
    main()
