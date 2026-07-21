import importlib.util
from pathlib import Path


def load():
    path = Path(__file__).parent / "main.py"
    spec = importlib.util.spec_from_file_location("svc", path)
    mod = importlib.util.module_from_spec(spec)
    assert spec.loader
    spec.loader.exec_module(mod)
    return mod


def test_demo_payload_offline_llm_proof():
    mod = load()
    payload = mod.demo_payload()

    assert payload["ok"] is True
    assert payload["offline_fixture_llm"] is True
    assert payload["offline_llm"] is True
    assert payload["fixture_mode"] is True
    assert payload["live_provider"] is False
    assert payload["simulator"] is True
    assert payload["summary"]
    assert payload["failures_ingested"] >= 1


def test_info_honesty_labels():
    mod = load()

    payload = mod.info()

    assert payload["offline_fixture_llm"] is True
    assert payload["live_provider"] is False
    assert payload["simulator"] is True
