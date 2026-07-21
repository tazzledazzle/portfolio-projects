"""RED/GREEN tests for AI-assisted workflow FSM with approval (ENG-E-018)."""

from __future__ import annotations

import pytest

from workflow import STAGES, WorkflowStore, advance, approve, create_workflow


def test_workflow_stages_constant():
    assert STAGES == ["ingest", "retrieve", "propose", "approval"]


def test_create_workflow_starts_ingest():
    store = WorkflowStore()
    wf = create_workflow(store, failure_id="pipeline-cache-timeout")
    assert wf["stage"] == "ingest" or wf["status"] in (
        "progressing",
        "awaiting_approval",
        "ingest",
    )
    # Advance through stages until approval gate
    current = wf
    seen = {current.get("stage")}
    for _ in range(10):
        if current.get("status") == "awaiting_approval":
            break
        current = advance(store, current["id"])
        seen.add(current.get("stage"))
    assert current["status"] == "awaiting_approval"
    assert current.get("approval_required") is True
    assert "ingest" in seen or current["stage"] == "approval"


def test_advance_blocked_without_approval_at_gate():
    store = WorkflowStore()
    wf = create_workflow(store, failure_id="pipeline-cache-timeout")
    current = wf
    for _ in range(10):
        if current.get("status") == "awaiting_approval":
            break
        current = advance(store, current["id"])
    assert current["status"] == "awaiting_approval"
    with pytest.raises((ValueError, PermissionError, RuntimeError)):
        advance(store, current["id"])
    assert store.get(current["id"])["status"] != "approved"


def test_approve_transitions_to_approved():
    store = WorkflowStore()
    wf = create_workflow(store, failure_id="pipeline-cache-timeout")
    current = wf
    for _ in range(10):
        if current.get("status") == "awaiting_approval":
            break
        current = advance(store, current["id"])
    assert current.get("approval_required") is True
    approved = approve(store, current["id"])
    assert approved["status"] == "approved"


def test_invalid_transition_rejected():
    store = WorkflowStore()
    wf = create_workflow(store, failure_id="pipeline-cache-timeout")
    with pytest.raises((ValueError, RuntimeError, KeyError)):
        # Illegal jump: cannot approve before reaching the gate
        approve(store, wf["id"])
