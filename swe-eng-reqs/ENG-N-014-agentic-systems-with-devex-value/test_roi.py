import json
from pathlib import Path

from roi import compute_roi
from value_feature import run_value_feature


def fixture_baseline():
    path = Path(__file__).parent / "testdata" / "baselines.json"
    return json.loads(path.read_text())["baseline"]


def test_compute_roi_from_fixtures():
    baseline = fixture_baseline()
    assisted = run_value_feature(baseline)

    result = compute_roi(baseline, assisted)

    assert isinstance(result["time_saved_minutes"], (int, float))
    assert result["time_saved_minutes"] > 0
    assert isinstance(result["mttr_improvement_pct"], (int, float))
    assert result["mttr_improvement_pct"] > 0


def test_roi_baseline_source_fixture():
    baseline = fixture_baseline()

    result = compute_roi(baseline, run_value_feature(baseline))

    assert result["baseline_source"] == "fixture"


def test_roi_fabricated_prod_false():
    baseline = fixture_baseline()

    result = compute_roi(baseline, run_value_feature(baseline))

    assert result["fabricated_prod"] is False


def test_roi_narrative_non_empty():
    baseline = fixture_baseline()

    result = compute_roi(baseline, run_value_feature(baseline))

    assert isinstance(result["narrative"], str)
    assert result["narrative"].strip()


def test_value_feature_produces_assisted_path():
    assisted = run_value_feature(fixture_baseline())

    assert assisted["source"] == "fixture_assisted"
    assert assisted["resolution_minutes"] > 0
    assert assisted["mttr_minutes"] > 0
