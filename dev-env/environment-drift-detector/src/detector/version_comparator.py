from dataclasses import dataclass
from enum import Enum


class DriftCategory(str, Enum):
    OK = "ok"
    OUTDATED = "outdated"
    AHEAD = "ahead"
    MISSING = "missing"


@dataclass(frozen=True)
class DriftResult:
    tool: str
    expected: str
    actual: str | None
    category: DriftCategory


def normalize_version(version: str) -> tuple[int, ...]:
    parts: list[int] = []
    for piece in version.strip().split("."):
        digits = "".join(character for character in piece if character.isdigit())
        parts.append(int(digits) if digits else 0)
    return tuple(parts)


def _pad_versions(
    left: tuple[int, ...],
    right: tuple[int, ...],
) -> tuple[tuple[int, ...], tuple[int, ...]]:
    width = max(len(left), len(right))
    return (
        left + (0,) * (width - len(left)),
        right + (0,) * (width - len(right)),
    )


def compare_tool_version(expected: str, actual: str | None) -> DriftCategory:
    if actual is None:
        return DriftCategory.MISSING
    expected_parts, actual_parts = _pad_versions(
        normalize_version(expected),
        normalize_version(actual),
    )
    if actual_parts < expected_parts:
        return DriftCategory.OUTDATED
    if actual_parts > expected_parts:
        return DriftCategory.AHEAD
    return DriftCategory.OK


def detect_drift(
    canonical: dict[str, str],
    local: dict[str, str | None],
) -> list[DriftResult]:
    results: list[DriftResult] = []
    for tool, expected in canonical.items():
        actual = local.get(tool)
        results.append(
            DriftResult(
                tool=tool,
                expected=expected,
                actual=actual,
                category=compare_tool_version(expected, actual),
            )
        )
    return results
