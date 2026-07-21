import importlib.util
from pathlib import Path


def load():
    path = Path(__file__).parent / "main.py"
    spec = importlib.util.spec_from_file_location("svc", path)
    mod = importlib.util.module_from_spec(spec)
    assert spec.loader
    spec.loader.exec_module(mod)
    return mod


def test_demo_payload_policy_gateway_proof():
    mod = load()
    payload = mod.demo_payload()

    assert payload["ok"] is True
    assert payload["mutate_denied_without_approval"] is True
    assert payload["mutate_allowed_with_approval"] is True
    assert payload["policy_gateway"] is True
    assert payload["mcp_inspired"] is True
    assert payload["mcp_sdk"] is False
    assert payload["audit_entries"] >= 2
