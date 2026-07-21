"""Deterministic agent-assisted path used to measure fixture DevEx value."""

from __future__ import annotations


def run_value_feature(baseline: dict) -> dict:
    """Run a thin offline diagnose/propose path over a fixture baseline."""
    resolution_minutes = float(baseline["resolution_minutes"])
    mttr_minutes = float(baseline["mttr_minutes"])
    trace = [
        {"stage": "diagnose", "result": "identified failing test cluster"},
        {"stage": "retrieve", "result": "loaded matching runbook fixture"},
        {"stage": "propose", "result": "suggested bounded retry and owner routing"},
    ]
    return {
        "source": "fixture_assisted",
        "resolution_minutes": round(resolution_minutes * 0.6, 2),
        "mttr_minutes": round(mttr_minutes * 0.65, 2),
        "trace": trace,
        "executed_mutation": False,
    }
