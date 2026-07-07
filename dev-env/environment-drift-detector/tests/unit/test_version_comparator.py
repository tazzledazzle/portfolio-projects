from pathlib import Path

from detector.version_comparator import DriftCategory, compare_tool_version, detect_drift
from fix_script.script_generator import generate_fix_script
from manifest.schema import load_canonical_policy, load_tool_versions


def test_compare_tool_version_detects_outdated() -> None:
    assert compare_tool_version("21.0.2", "21.0.1") == DriftCategory.OUTDATED


def test_detect_drift_from_fixture_files() -> None:
    root = Path(__file__).resolve().parents[2]
    canonical = load_canonical_policy(root / "config/canonical-toolchain.yaml")
    local = load_tool_versions(root / "tests/fixtures/manifests/sample-manifest.yaml")
    drifts = detect_drift(canonical, local)
    drift_by_tool = {drift.tool: drift for drift in drifts}
    assert drift_by_tool["node"].category == DriftCategory.OK
    assert drift_by_tool["jvm"].category == DriftCategory.OUTDATED
    assert drift_by_tool["kubectl"].category == DriftCategory.MISSING


def test_generate_fix_script_includes_outdated_tools() -> None:
    drifts = detect_drift({"jvm": "21"}, {"jvm": "17"})
    script = generate_fix_script(drifts)
    assert "Update jvm to 21" in script
