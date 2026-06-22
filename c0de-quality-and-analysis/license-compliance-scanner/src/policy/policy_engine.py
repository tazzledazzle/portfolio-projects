from dataclasses import dataclass


@dataclass
class Violation:
    package: str
    license_id: str
    reason: str


def evaluate_policy(packages: list[dict], allowed: set[str], restricted: set[str]) -> list[Violation]:
    violations: list[Violation] = []
    for package in packages:
        license_id = package.get("license", "UNKNOWN")
        if license_id in restricted or (allowed and license_id not in allowed):
            violations.append(
                Violation(
                    package=package.get("name", "unknown"),
                    license_id=license_id,
                    reason="License is not compliant with policy.",
                )
            )
    return violations
