"""RED/GREEN tests for propose-only safe actions (ENG-E-006)."""

from __future__ import annotations

import copy
from pathlib import Path

from diagnose import diagnose_pipeline, load_pipeline_fixtures
from propose import propose_actions

ROOT = Path(__file__).resolve().parent
FIXTURES = ROOT / "testdata" / "pipelines.json"


def test_propose_actions_marks_mutating_requires_approval():
    fixtures = load_pipeline_fixtures(FIXTURES)
    diagnosis = diagnose_pipeline(fixtures[0]["id"], fixtures_path=FIXTURES)
    result = propose_actions(diagnosis)
    actions = result.get("proposed_actions") if isinstance(result, dict) else result
    assert actions and len(actions) >= 1
    mutating = [a for a in actions if a.get("mutating")]
    assert mutating, "expected at least one mutating proposed action"
    assert all(a.get("requires_approval") is True for a in mutating)


def test_propose_never_executes():
    fixtures = load_pipeline_fixtures(FIXTURES)
    before = copy.deepcopy(fixtures)
    diagnosis = diagnose_pipeline(fixtures[0]["id"], fixtures_path=FIXTURES)
    result = propose_actions(diagnosis, fixtures=fixtures)
    assert result["executed"] is False
    after = load_pipeline_fixtures(FIXTURES)
    assert after == before
    # In-memory fixture list must not be mutated either
    assert fixtures == before
