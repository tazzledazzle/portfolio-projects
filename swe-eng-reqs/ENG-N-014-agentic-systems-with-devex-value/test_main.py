import importlib.util
from pathlib import Path


def load():
    path = Path(__file__).parent / "main.py"
    spec = importlib.util.spec_from_file_location("svc", path)
    mod = importlib.util.module_from_spec(spec)
    assert spec.loader
    spec.loader.exec_module(mod)
    return mod


def test_demo_payload_roi_proof():
    mod = load()
    payload = mod.demo_payload()

    assert payload["ok"] is True
    assert payload["time_saved_minutes"] > 0
    assert payload["mttr_improvement_pct"] > 0
    assert payload["baseline_source"] == "fixture"
    assert payload["fabricated_prod"] is False
    assert payload["narrative"].strip()
