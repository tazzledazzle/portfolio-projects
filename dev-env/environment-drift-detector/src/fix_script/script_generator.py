from detector.version_comparator import DriftCategory, DriftResult


def generate_fix_script(drifts: list[DriftResult]) -> str:
    lines = ["#!/usr/bin/env bash", "set -euo pipefail", ""]
    for drift in drifts:
        if drift.category == DriftCategory.OUTDATED:
            lines.append(
                f'echo "Update {drift.tool} to {drift.expected} (found {drift.actual})"'
            )
        elif drift.category == DriftCategory.MISSING:
            lines.append(f'echo "Install {drift.tool} at version {drift.expected}"')
    if len(lines) == 3:
        lines.append('echo "No remediation required."')
    return "\n".join(lines) + "\n"
