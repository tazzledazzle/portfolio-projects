from detector.version_comparator import DriftCategory, DriftResult


def format_drift_report(drifts: list[DriftResult]) -> str:
    lines = ["Environment drift report", "========================", ""]
    for drift in drifts:
        actual = drift.actual if drift.actual is not None else "missing"
        lines.append(
            f"- {drift.tool}: expected {drift.expected}, actual {actual} "
            f"({drift.category.value})"
        )
    outdated = sum(1 for drift in drifts if drift.category != DriftCategory.OK)
    lines.append("")
    lines.append(f"Summary: {outdated} tool(s) need attention.")
    return "\n".join(lines) + "\n"
