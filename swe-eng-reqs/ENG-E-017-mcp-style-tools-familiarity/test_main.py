import importlib.util
from pathlib import Path


def load():
    path = Path(__file__).parent / "main.py"
    spec = importlib.util.spec_from_file_location("svc", path)
    mod = importlib.util.module_from_spec(spec)
    assert spec.loader
    spec.loader.exec_module(mod)
    return mod


def test_demo_payload_mcp_proof():
    mod = load()
    payload = mod.demo_payload()

    assert payload["ok"] is True
    assert payload["mcp_inspired"] is True
    assert payload["mcp_sdk"] is False
    assert payload["tools_listed"] >= 2
    assert payload["auth_ok"] is True
    assert payload["audit_entries"] >= 1
    assert payload["read_only"] is True


def test_info_has_honesty_labels():
    mod = load()

    assert mod.info()["mcp_inspired"] is True
    assert mod.info()["mcp_sdk"] is False
    assert mod.info()["simulator"] is True
