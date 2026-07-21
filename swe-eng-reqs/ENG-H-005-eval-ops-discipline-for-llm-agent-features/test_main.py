import importlib.util
from pathlib import Path


def load():
    path = Path(__file__).parent / "main.py"
    spec = importlib.util.spec_from_file_location("svc", path)
    mod = importlib.util.module_from_spec(spec)
    assert spec.loader
    spec.loader.exec_module(mod)
    return mod


def test_demo_payload_eval_ops_proof():
    mod = load()
    payload = mod.demo_payload()

    assert payload["ok"] is True
    assert payload["offline_pass"] is True
    assert payload["online_sim_pass"] is True
    assert payload["failure_fixtures_caught"] >= 1
    assert payload["eval_ops"] is True
    assert "offline" in payload["mode"]
    assert payload["simulator"] is True


def test_info_has_eval_honesty_labels():
    mod = load()

    assert mod.info()["eval_ops"] is True
    assert mod.info()["simulator"] is True
    assert mod.info()["live_provider"] is False
