"""Fixture-only time-saved and MTTR value calculations."""

from __future__ import annotations


def compute_roi(baseline: dict, assisted: dict) -> dict:
    baseline_resolution = float(baseline["resolution_minutes"])
    baseline_mttr = float(baseline["mttr_minutes"])
    assisted_resolution = float(assisted["resolution_minutes"])
    assisted_mttr = float(assisted["mttr_minutes"])
    if min(baseline_resolution, baseline_mttr, assisted_resolution, assisted_mttr) <= 0:
        raise ValueError("fixture durations must be positive")

    time_saved = round(baseline_resolution - assisted_resolution, 2)
    mttr_improvement = round(
        ((baseline_mttr - assisted_mttr) / baseline_mttr) * 100,
        2,
    )
    narrative = (
        f"In this deterministic fixture, the assisted path saves {time_saved:g} "
        f"minutes and improves MTTR by {mttr_improvement:g}%. These are "
        "scenario results, not production telemetry."
    )
    return {
        "time_saved_minutes": time_saved,
        "mttr_improvement_pct": mttr_improvement,
        "baseline_source": "fixture",
        "fabricated_prod": False,
        "narrative": narrative,
    }
