from pathlib import Path
import sys

sys.path.append(str(Path(__file__).resolve().parents[1] / "src"))

from main import build_spans


def test_build_spans_generates_one_span_per_step() -> None:
    spans = build_spans(
        {
            "job": "lint",
            "steps": [
                {"name": "ruff", "status": "passed"},
                {"name": "mypy", "status": "failed", "failure_reason": "type error"},
            ],
        }
    )
    assert len(spans) == 2
    assert spans[1]["attributes"]["failure_reason"] == "type error"
