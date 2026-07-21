"""Offline and local online-simulation eval runners."""

from __future__ import annotations

import json
from pathlib import Path
from typing import Any, Callable

from scorer import ScoreReport, score_case

LocalRunner = Callable[[dict[str, Any]], str]


class EvalHarness:
    def __init__(self, fixture_dir: Path | str) -> None:
        self.fixture_dir = Path(fixture_dir)
        self.cases = self._load_cases()

    def _load_cases(self) -> list[dict[str, Any]]:
        golden = json.loads((self.fixture_dir / "evals.json").read_text())
        known_bad = json.loads(
            (self.fixture_dir / "injection_bad.json").read_text()
        )
        if not isinstance(golden, list) or not isinstance(known_bad, dict):
            raise ValueError("eval fixtures have invalid shape")
        return [*golden, known_bad]

    def run_offline(self) -> dict[str, Any]:
        return self._run(
            mode="offline",
            runner=lambda case: str(case["candidate"]),
            simulator=False,
        )

    def run_online_sim(self, runner: LocalRunner) -> dict[str, Any]:
        return self._run(mode="online-sim", runner=runner, simulator=True)

    def _run(
        self,
        *,
        mode: str,
        runner: LocalRunner,
        simulator: bool,
    ) -> dict[str, Any]:
        reports: list[ScoreReport] = []
        for case in self.cases:
            reports.append(score_case(case, runner(dict(case))))

        golden_pass = all(
            report.passed
            for report in reports
            if report.expected == "pass"
        )
        known_bad = [
            report for report in reports if report.expected == "fail"
        ]
        failures_caught = sum(report.caught for report in known_bad)
        suite_pass = golden_pass and bool(known_bad) and all(
            report.caught and not report.passed for report in known_bad
        )
        return {
            f"{mode.replace('-', '_')}_pass": suite_pass,
            "mode": mode,
            "simulator": simulator,
            "cases_run": len(reports),
            "golden_passed": sum(
                report.passed
                for report in reports
                if report.expected == "pass"
            ),
            "failure_fixtures_caught": failures_caught,
            "reports": [report.to_dict() for report in reports],
        }


def run_offline(fixture_dir: Path | str) -> dict[str, Any]:
    return EvalHarness(fixture_dir).run_offline()


def run_online_sim(
    fixture_dir: Path | str,
    runner: LocalRunner,
) -> dict[str, Any]:
    return EvalHarness(fixture_dir).run_online_sim(runner)
