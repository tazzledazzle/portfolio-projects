#!/usr/bin/env python3
"""ENG-E-018: AI-assisted developer workflows

Multi-stage ingest→retrieve→propose→approval workflow with human gate.
Offline deterministic fixtures — no live LLM or API keys.
"""

from __future__ import annotations

import json
import os
from http.server import BaseHTTPRequestHandler, HTTPServer
from pathlib import Path
from urllib.parse import urlparse

from workflow import (
    STAGES,
    WorkflowStore,
    advance,
    approve,
    create_workflow,
    run_to_awaiting,
)

ROOT = Path(__file__).resolve().parent
STATE = {"runs": 0, "audits": [], "store": WorkflowStore()}


def _default_failure_id() -> str:
    path = ROOT / "testdata" / "failures.json"
    raw = json.loads(path.read_text(encoding="utf-8"))
    failures = raw.get("failures") or []
    if not failures:
        raise ValueError("testdata/failures.json must contain at least one failure")
    return str(failures[0]["id"])


def info() -> dict:
    return {
        "requirement_id": "ENG-E-018",
        "service": "eng-e-018",
        "title": "AI-assisted developer workflows",
        "stages": list(STAGES),
        "simulator": True,
        "live_provider": False,
        "offline_fixture_llm": False,
        "note": "Workflow FSM with approval gate; not OfflineFixtureLLM / ToolRegistry / MCP.",
    }


def demo_payload() -> dict:
    """Dual-path proof: run to awaiting_approval, then approve → approved."""
    STATE["runs"] += 1
    store = WorkflowStore()
    failure_id = _default_failure_id()
    gated = run_to_awaiting(store, failure_id)
    status_after_gate = gated["status"]
    awaiting_before_approve = status_after_gate == "awaiting_approval"
    approved = approve(store, gated["id"])
    STATE["audits"].append(
        {
            "action": "workflow_dual_path",
            "requirement_id": "ENG-E-018",
            "workflow_id": approved["id"],
            "status_after_gate": status_after_gate,
            "final_status": approved["status"],
        }
    )
    return {
        "ok": True,
        "requirement_id": "ENG-E-018",
        "service": "eng-e-018",
        "title": "AI-assisted developer workflows",
        "run": STATE["runs"],
        "simulator": True,
        "live_provider": False,
        "stages": list(STAGES),
        "workflow_id": approved["id"],
        "failure_id": failure_id,
        "approval_required": True,
        "awaiting_before_approve": awaiting_before_approve,
        "status_after_gate": status_after_gate,
        "status": approved["status"],
        "proposal": approved.get("proposal"),
        "history": approved.get("history"),
        "dual_path": "awaiting_then_approved",
        "path": ["awaiting_approval", "approved"],
        "audit_tail": STATE["audits"][-3:],
        "acceptance": [
            "healthz returns ok",
            "stages are ingest→retrieve→propose→approval",
            "workflow reaches awaiting_approval before approve",
            "approve transitions to approved",
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
            body = f"eng-e-018_demo_runs_total {STATE['runs']}\n".encode()
            self.send_response(200)
            self.send_header("Content-Type", "text/plain")
            self.end_headers()
            self.wfile.write(body)
            return
        self._json(404, {"error": "not found"})

    def do_POST(self) -> None:  # noqa: N802
        parsed = urlparse(self.path)
        path = parsed.path
        if path == "/v1/demo":
            self._json(200, demo_payload())
            return
        if path == "/v1/workflows":
            try:
                payload = self._request_json()
                failure_id = payload.get("failure_id") or _default_failure_id()
                if not isinstance(failure_id, str):
                    raise ValueError("failure_id must be a string")
                auto_advance = bool(payload.get("auto_advance", True))
                store: WorkflowStore = STATE["store"]
                if auto_advance:
                    wf = run_to_awaiting(store, failure_id)
                else:
                    wf = create_workflow(store, failure_id=failure_id)
                self._json(200, {"ok": True, "workflow": wf, "stages": list(STAGES)})
            except (json.JSONDecodeError, ValueError, RuntimeError) as exc:
                self._json(400, {"ok": False, "error": str(exc)})
            return
        if path.startswith("/v1/workflows/") and path.endswith("/approve"):
            try:
                workflow_id = path[len("/v1/workflows/") : -len("/approve")]
                if not workflow_id or "/" in workflow_id:
                    raise ValueError("invalid workflow id")
                wf = approve(STATE["store"], workflow_id)
                self._json(200, {"ok": True, "workflow": wf})
            except KeyError as exc:
                self._json(404, {"ok": False, "error": str(exc)})
            except (ValueError, PermissionError) as exc:
                self._json(400, {"ok": False, "error": str(exc)})
            return
        if path.startswith("/v1/workflows/") and path.endswith("/advance"):
            try:
                workflow_id = path[len("/v1/workflows/") : -len("/advance")]
                if not workflow_id or "/" in workflow_id:
                    raise ValueError("invalid workflow id")
                wf = advance(STATE["store"], workflow_id)
                self._json(200, {"ok": True, "workflow": wf})
            except KeyError as exc:
                self._json(404, {"ok": False, "error": str(exc)})
            except (ValueError, PermissionError) as exc:
                self._json(400, {"ok": False, "error": str(exc)})
            return
        self._json(404, {"error": "not found"})

    def log_message(self, fmt: str, *args) -> None:
        return


def main() -> None:
    addr = os.environ.get("ADDR", "0.0.0.0")
    port = int(os.environ.get("PORT", "8080"))
    server = HTTPServer((addr, port), Handler)
    print(f"eng-e-018 listening on {addr}:{port} (requirement ENG-E-018)", flush=True)
    server.serve_forever()


if __name__ == "__main__":
    main()
