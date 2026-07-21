"""Propose-only safe DevEx actions — never execute (ENG-E-006 / D-10)."""

from __future__ import annotations

from typing import Any


def propose_actions(
    diagnosis: dict[str, Any] | str,
    *,
    fixtures: list[dict[str, Any]] | None = None,
) -> dict[str, Any]:
    """Propose safe remediation actions without executing any mutation.

    Mutating proposals always set requires_approval=True and executed=False.
    The optional fixtures list is never mutated (propose-only contract).
    """
    if fixtures is not None:
        # Touch length only — never mutate fixture state.
        _ = len(fixtures)

    if isinstance(diagnosis, str):
        pipeline_id = "unknown"
        summary = diagnosis
        stage = "unknown"
        findings: list[str] = [diagnosis]
    else:
        pipeline_id = str(diagnosis.get("pipeline_id") or "unknown")
        summary = str(diagnosis.get("summary") or diagnosis.get("diagnosis") or "")
        stage = str(diagnosis.get("failed_stage") or "unknown")
        findings = list(diagnosis.get("findings") or [])

    actions: list[dict[str, Any]] = [
        {
            "name": "inspect_logs",
            "mutating": False,
            "requires_approval": False,
            "description": f"Read recent logs for pipeline {pipeline_id} stage {stage}",
        },
        {
            "name": "retry_failed_stage",
            "mutating": True,
            "requires_approval": True,
            "description": f"Retry stage '{stage}' for pipeline {pipeline_id}",
        },
        {
            "name": "propose_promote",
            "mutating": True,
            "requires_approval": True,
            "description": f"Propose promote only after {pipeline_id} is healthy",
        },
    ]

    mutating = [a for a in actions if a.get("mutating")]
    all_mutating_require_approval = bool(mutating) and all(
        a.get("requires_approval") is True for a in mutating
    )

    return {
        "pipeline_id": pipeline_id,
        "summary": summary,
        "findings": findings,
        "proposed_actions": actions,
        "all_mutating_require_approval": all_mutating_require_approval,
        "executed": False,
        "execute_mutating": False,
        "note": "Propose-only; mutating actions require human approval (N-015 owns grants).",
    }
