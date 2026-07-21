#!/usr/bin/env python3
"""ENG-E-006: Agentic AI tooling for DevEx

Diagnose pipeline fixtures and propose safe actions without executing them.
Offline deterministic fixtures only — no live LLM or API keys.
"""

from __future__ import annotations

import json
import os
from http.server import BaseHTTPRequestHandler, HTTPServer
from pathlib import Path

from diagnose import diagnose_pipeline, load_pipeline_fixtures
from propose import propose_actions

ROOT = Path(__file__).resolve().parent
STATE = {"runs": 0, "audits": []}


def info() -> dict:
    return {
        "requirement_id": "ENG-E-006",
        "service": "eng-e-006",
        "title": "Agentic AI tooling for DevEx",
        "offline_fixture_llm": True,
        "simulator": True,
        "live_provider": False,
        "execute_mutating": False,
        "note": "Diagnose + propose-only; mutating actions require approval; no policy gateway.",
    }


def run_diagnose(pipeline_id: str | None = None) -> dict:
    fixtures = load_pipeline_fixtures(ROOT / "testdata" / "pipelines.json")
    target = pipeline_id or fixtures[0]["id"]
    diagnosis = diagnose_pipeline(target, fixtures=fixtures)
    proposed = propose_actions(diagnosis, fixtures=fixtures)
    return {
        "diagnosis": diagnosis,
        "proposed_actions": proposed["proposed_actions"],
        "all_mutating_require_approval": proposed["all_mutating_require_approval"],
        "executed": proposed["executed"],
        "pipeline_id": target,
        "summary": diagnosis.get("summary") or diagnosis.get("diagnosis"),
    }


def demo_payload() -> dict:
    STATE["runs"] += 1
    result = run_diagnose()
    STATE["audits"].append(
        {
            "action": "diagnose_propose",
            "requirement_id": "ENG-E-006",
            "pipeline_id": result["pipeline_id"],
            "executed": False,
        }
    )
    return {
        "ok": True,
        "requirement_id": "ENG-E-006",
        "service": "eng-e-006",
        "title": "Agentic AI tooling for DevEx",
        "run": STATE["runs"],
        "offline_fixture_llm": True,
        "simulator": True,
        "live_provider": False,
        "execute_mutating": False,
        "diagnosis": result["summary"],
        "diagnosis_detail": result["diagnosis"],
        "proposed_actions": result["proposed_actions"],
        "all_mutating_require_approval": result["all_mutating_require_approval"],
        "executed": result["executed"],
        "pipeline_id": result["pipeline_id"],
        "audit_tail": STATE["audits"][-3:],
        "acceptance": [
            "healthz returns ok",
            "pipeline fixtures produce non-empty diagnosis",
            "mutating proposed actions require approval",
            "demo never executes mutating actions",
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
            body = f"eng-e-006_demo_runs_total {STATE['runs']}\n".encode()
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
        if self.path == "/v1/diagnose":
            try:
                payload = self._request_json()
                pipeline_id = payload.get("pipeline_id")
                if pipeline_id is not None and not isinstance(pipeline_id, str):
                    raise ValueError("pipeline_id must be a string")
                result = run_diagnose(pipeline_id)
                self._json(200, {"ok": True, **result, "executed": False})
            except KeyError as exc:
                self._json(404, {"ok": False, "error": str(exc)})
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
    print(f"eng-e-006 listening on {addr}:{port} (requirement ENG-E-006)", flush=True)
    server.serve_forever()


if __name__ == "__main__":
    main()
