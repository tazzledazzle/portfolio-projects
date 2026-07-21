import json
from pathlib import Path

from scorer import ScoreReport, score_case

TESTDATA = Path(__file__).parent / "testdata"


def load_json(name: str):
    return json.loads((TESTDATA / name).read_text())


def test_scorer_golden_passes():
    case = load_json("evals.json")[0]

    report = score_case(case, case["candidate"])

    assert isinstance(report, ScoreReport)
    assert report.passed is True
    assert report.caught is False


def test_scorer_known_bad_fails():
    case = load_json("injection_bad.json")

    report = score_case(case, case["candidate"])

    assert report.passed is False
    assert report.caught is True
    assert "prompt_injection" in report.reasons


def test_scorer_injection_fixture_caught():
    case = load_json("injection_bad.json")
    report = score_case(case, case["candidate"])

    assert case["expected"] == "fail"
    assert report.caught is True
    assert report.risk == "injection"
