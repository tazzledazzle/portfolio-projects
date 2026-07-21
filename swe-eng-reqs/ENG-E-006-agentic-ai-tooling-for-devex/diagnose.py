"""Pipeline fixture diagnosis for DevEx agent tooling (ENG-E-006).

Deterministic offline diagnosis from on-disk fixtures — no live LLM.
"""

from __future__ import annotations

import json
from pathlib import Path
from typing import Any

ROOT = Path(__file__).resolve().parent
DEFAULT_FIXTURES = ROOT / "testdata" / "pipelines.json"


def load_pipeline_fixtures(path: Path | str | None = None) -> list[dict[str, Any]]:
    fixtures_path = Path(path) if path is not None else DEFAULT_FIXTURES
    raw = json.loads(fixtures_path.read_text(encoding="utf-8"))
    pipelines = raw.get("pipelines") if isinstance(raw, dict) else raw
    if not isinstance(pipelines, list) or not pipelines:
        raise ValueError("pipeline fixtures must be a non-empty list")
    for item in pipelines:
        if not isinstance(item, dict) or not item.get("id"):
            raise ValueError("each pipeline fixture requires an id")
    return pipelines


def diagnose_pipeline(
    pipeline_id: str,
    *,
    fixtures_path: Path | str | None = None,
    fixtures: list[dict[str, Any]] | None = None,
) -> dict[str, Any]:
    """Return a structured diagnosis for a known pipeline fixture id."""
    items = fixtures if fixtures is not None else load_pipeline_fixtures(fixtures_path)
    match = next((p for p in items if p.get("id") == pipeline_id), None)
    if match is None:
        raise KeyError(f"unknown pipeline id: {pipeline_id}")

    status = match.get("status", "unknown")
    stage = match.get("failed_stage") or match.get("stage") or "unknown"
    error = match.get("error") or match.get("message") or "unspecified failure"
    findings = list(match.get("findings") or [])
    if not findings:
        findings = [f"pipeline {pipeline_id} reported status={status} at stage={stage}"]

    summary = (
        f"Pipeline {pipeline_id} is {status} at stage '{stage}': {error}. "
        f"Recommended focus: {findings[0]}."
    )
    return {
        "pipeline_id": pipeline_id,
        "status": status,
        "failed_stage": stage,
        "error": error,
        "findings": findings,
        "summary": summary,
        "diagnosis": summary,
        "offline_fixture_llm": True,
        "live_provider": False,
    }
