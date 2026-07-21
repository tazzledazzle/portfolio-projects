import importlib.util
from pathlib import Path


def load():
    path = Path(__file__).parent / "main.py"
    spec = importlib.util.spec_from_file_location("svc", path)
    mod = importlib.util.module_from_spec(spec)
    assert spec.loader
    spec.loader.exec_module(mod)
    return mod


def test_demo_payload_devex_agent_proof():
    mod = load()
    payload = mod.demo_payload()
    assert payload["ok"] is True
    assert payload.get("diagnosis")
    actions = payload.get("proposed_actions") or []
    assert len(actions) >= 1
    assert payload.get("all_mutating_require_approval") is True
    assert payload.get("executed") is False
    mutating = [a for a in actions if a.get("mutating")]
    assert mutating
    assert all(a.get("requires_approval") is True for a in mutating)


def test_info_honesty_labels():
    mod = load()
    payload = mod.info()
    assert payload.get("offline_fixture_llm") is True or payload.get("simulator") is True
    assert payload.get("live_provider") is False
    assert payload.get("execute_mutating") is False
