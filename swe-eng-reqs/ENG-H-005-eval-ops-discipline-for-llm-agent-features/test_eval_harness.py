from pathlib import Path

from eval_harness import EvalHarness, run_offline, run_online_sim

TESTDATA = Path(__file__).parent / "testdata"


def test_harness_offline_pass():
    result = run_offline(TESTDATA)

    assert result["offline_pass"] is True
    assert result["mode"] == "offline"
    assert result["cases_run"] >= 2


def test_harness_online_sim_pass():
    calls = []

    def local_stub(case):
        calls.append(case["id"])
        return case["candidate"]

    result = run_online_sim(TESTDATA, local_stub)

    assert result["online_sim_pass"] is True
    assert result["simulator"] is True
    assert len(calls) == result["cases_run"]


def test_harness_failure_fixtures_caught_count():
    harness = EvalHarness(TESTDATA)
    result = harness.run_offline()

    assert result["failure_fixtures_caught"] >= 1
    bad_report = next(report for report in result["reports"] if report["expected"] == "fail")
    assert bad_report["passed"] is False
    assert bad_report["caught"] is True
