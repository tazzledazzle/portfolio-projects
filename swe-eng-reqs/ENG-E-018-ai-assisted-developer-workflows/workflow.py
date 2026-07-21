"""Multi-stage AI-assisted developer workflow with approval gate (ENG-E-018).

Stages: ingest → retrieve → propose → approval (awaiting_approval → approved).
"""

from __future__ import annotations

import itertools
import threading
from typing import Any

STAGES = ["ingest", "retrieve", "propose", "approval"]

# Status values used by the FSM
STATUS_PROGRESSING = "progressing"
STATUS_AWAITING = "awaiting_approval"
STATUS_APPROVED = "approved"


class WorkflowStore:
    """In-memory mutex-backed workflow store."""

    def __init__(self) -> None:
        self._lock = threading.Lock()
        self._items: dict[str, dict[str, Any]] = {}
        self._seq = itertools.count(1)

    def put(self, wf: dict[str, Any]) -> dict[str, Any]:
        with self._lock:
            self._items[wf["id"]] = dict(wf)
            return dict(self._items[wf["id"]])

    def get(self, workflow_id: str) -> dict[str, Any]:
        with self._lock:
            if workflow_id not in self._items:
                raise KeyError(f"unknown workflow id: {workflow_id}")
            return dict(self._items[workflow_id])

    def next_id(self) -> str:
        with self._lock:
            return f"wf-{next(self._seq)}"


def create_workflow(
    store: WorkflowStore,
    *,
    failure_id: str,
    context: dict[str, Any] | None = None,
) -> dict[str, Any]:
    if not failure_id or not isinstance(failure_id, str):
        raise ValueError("failure_id must be a non-empty string")
    wf = {
        "id": store.next_id(),
        "failure_id": failure_id,
        "stage": "ingest",
        "status": STATUS_PROGRESSING,
        "approval_required": True,
        "proposal": None,
        "context": dict(context or {}),
        "history": ["ingest"],
    }
    return store.put(wf)


def advance(store: WorkflowStore, workflow_id: str) -> dict[str, Any]:
    wf = store.get(workflow_id)
    status = wf["status"]
    if status == STATUS_APPROVED:
        raise ValueError("workflow already approved; no further advances")
    if status == STATUS_AWAITING:
        raise PermissionError(
            "advance blocked at approval gate; call approve() to continue"
        )

    stage = wf["stage"]
    if stage not in STAGES:
        raise ValueError(f"unknown stage: {stage}")

    idx = STAGES.index(stage)
    if stage == "ingest":
        wf["context"]["ingested"] = True
        wf["context"]["failure_id"] = wf["failure_id"]
        wf["stage"] = "retrieve"
        wf["history"].append("retrieve")
    elif stage == "retrieve":
        wf["context"]["retrieved"] = True
        wf["context"]["snippets"] = [
            f"fixture context for {wf['failure_id']}",
        ]
        wf["stage"] = "propose"
        wf["history"].append("propose")
    elif stage == "propose":
        wf["proposal"] = {
            "action": "retry_failed_stage",
            "mutating": True,
            "requires_approval": True,
            "summary": f"Proposed remediation for {wf['failure_id']}",
        }
        wf["stage"] = "approval"
        wf["status"] = STATUS_AWAITING
        wf["approval_required"] = True
        wf["history"].append("approval")
    elif stage == "approval":
        # Already at gate — should have been caught by status check
        raise PermissionError(
            "advance blocked at approval gate; call approve() to continue"
        )
    else:
        raise ValueError(f"illegal advance from stage {stage} (index {idx})")

    return store.put(wf)


def approve(store: WorkflowStore, workflow_id: str) -> dict[str, Any]:
    wf = store.get(workflow_id)
    if wf["status"] != STATUS_AWAITING:
        raise ValueError(
            f"cannot approve from status={wf['status']!r}; "
            "workflow must be awaiting_approval"
        )
    if wf["stage"] != "approval":
        raise ValueError(f"cannot approve from stage={wf['stage']!r}")
    wf["status"] = STATUS_APPROVED
    wf["approval_required"] = True  # was required; now granted
    wf["history"].append("approved")
    return store.put(wf)


def run_to_awaiting(store: WorkflowStore, failure_id: str) -> dict[str, Any]:
    """Create and advance until awaiting_approval."""
    wf = create_workflow(store, failure_id=failure_id)
    for _ in range(len(STAGES) + 2):
        if wf["status"] == STATUS_AWAITING:
            return wf
        wf = advance(store, wf["id"])
    raise RuntimeError("failed to reach awaiting_approval")
