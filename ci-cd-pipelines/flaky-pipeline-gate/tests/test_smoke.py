from pathlib import Path
import sys

sys.path.append(str(Path(__file__).resolve().parents[1] / "src"))

from main import gate_decision


def test_gate_allows_likely_flaky_failures() -> None:
    result = gate_decision(["fail", "fail", "pass", "fail"], threshold=0.5)
    assert result["decision"] == "allow"


def test_gate_blocks_likely_regressions() -> None:
    result = gate_decision(["pass", "pass", "fail", "pass"], threshold=0.5)
    assert result["decision"] == "block"
