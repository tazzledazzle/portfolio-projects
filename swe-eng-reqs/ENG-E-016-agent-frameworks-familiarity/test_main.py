import importlib.util
from pathlib import Path


def load():
    path = Path(__file__).parent / "main.py"
    spec = importlib.util.spec_from_file_location("svc", path)
    mod = importlib.util.module_from_spec(spec)
    assert spec.loader
    spec.loader.exec_module(mod)
    return mod


def test_demo_payload_agent_framework_proof():
    mod = load()
    payload = mod.demo_payload()

    assert payload["ok"] is True
    assert payload["agent_framework_inspired"] is True
    assert payload["simulator"] is True
    assert payload["live_provider"] is False
    assert payload["plan_steps"] >= 1
    assert payload["tools_invoked"] >= 1
    assert payload["deterministic_eval_pass"] is True
    assert payload["max_steps"] <= 5


def test_info_agent_framework_inspired():
    mod = load()

    payload = mod.info()

    assert payload["agent_framework_inspired"] is True
    assert payload["simulator"] is True
    assert payload["live_provider"] is False
