from dataclasses import dataclass


@dataclass(frozen=True)
class RiskResult:
    score: int
    reasons: list[str]
    approval_required: bool
    blocked: bool


def score_write_action(target_path: str, profile: str, dry_run: bool) -> tuple[int, list[str]]:
    reasons: list[str] = []
    score = 5
    if not dry_run:
        score += 20
        reasons.append("writes_file")
    if profile == "full-access":
        score += 30
        reasons.append("full_access_profile")
    if "/../" in target_path or target_path.startswith("../"):
        score += 40
        reasons.append("path_escape_pattern")
    return min(score, 100), reasons


def evaluate_risk(
    score: int, reasons: list[str], auto_allow_max: int, approval_required_min: int, hard_block_min: int
) -> RiskResult:
    approval_required = score >= approval_required_min
    blocked = score >= hard_block_min
    return RiskResult(score=score, reasons=reasons, approval_required=approval_required, blocked=blocked)
