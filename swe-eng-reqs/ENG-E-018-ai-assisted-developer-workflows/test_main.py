import importlib.util
from pathlib import Path


def load():
    path = Path(__file__).parent / "main.py"
    spec = importlib.util.spec_from_file_location("svc", path)
    mod = importlib.util.module_from_spec(spec)
    assert spec.loader
    spec.loader.exec_module(mod)
    return mod


def test_demo_payload_workflow_proof():
    mod = load()
    payload = mod.demo_payload()
    assert payload["ok"] is True
    assert payload["stages"] == ["ingest", "retrieve", "propose", "approval"]
    assert payload["approval_required"] is True
    assert payload["status"] in ("awaiting_approval", "approved")
    # Dual-path proof (N-006 style): reach gate, then approve
    assert payload.get("awaiting_before_approve") is True
    assert payload.get("status_after_gate") == "awaiting_approval"
    assert payload["status"] == "approved"
