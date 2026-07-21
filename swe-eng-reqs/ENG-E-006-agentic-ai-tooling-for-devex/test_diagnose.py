"""RED/GREEN tests for pipeline fixture diagnosis (ENG-E-006)."""

from __future__ import annotations

from pathlib import Path

import pytest

from diagnose import diagnose_pipeline, load_pipeline_fixtures

ROOT = Path(__file__).resolve().parent
FIXTURES = ROOT / "testdata" / "pipelines.json"


def test_diagnose_pipeline_from_fixtures():
    fixtures = load_pipeline_fixtures(FIXTURES)
    assert fixtures
    first_id = fixtures[0]["id"]
    diagnosis = diagnose_pipeline(first_id, fixtures_path=FIXTURES)
    assert diagnosis
    if isinstance(diagnosis, dict):
        assert diagnosis.get("summary") or diagnosis.get("diagnosis")
    else:
        assert isinstance(diagnosis, str) and diagnosis.strip()


def test_diagnose_unknown_pipeline_errors():
    with pytest.raises((KeyError, ValueError, LookupError)):
        diagnose_pipeline("pipeline-does-not-exist", fixtures_path=FIXTURES)
