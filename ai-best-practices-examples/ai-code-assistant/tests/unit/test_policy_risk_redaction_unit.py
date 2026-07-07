from pathlib import Path

from ai_code_assistant.policy import load_policy
from ai_code_assistant.redaction import redact_mapping, redact_text
from ai_code_assistant.risk import evaluate_risk, score_write_action


def test_load_policy_from_file(tmp_path: Path) -> None:
    policy_file = tmp_path / "assistant-policy.toml"
    policy_file.write_text(
        "[risk]\nauto_allow_max = 10\napproval_required_min = 20\nhard_block_min = 60\n\n[redaction]\nenabled = true\npatterns = ['token_[A-Za-z0-9]+']\n",
        encoding="utf-8",
    )
    policy, source = load_policy(str(policy_file))
    assert source == str(policy_file)
    assert policy.hard_block_min == 60


def test_redaction_masks_tokens_and_sensitive_keys() -> None:
    text = redact_text("my sk-ABCDEFGHIJKLMNOP is here", [r"sk-[A-Za-z0-9]+"])
    assert "***REDACTED***" in text
    mapped = redact_mapping({"api_key": "secret", "note": "ghp_12345678901234567890"}, [r"ghp_[A-Za-z0-9]+"])
    assert mapped["api_key"] == "***REDACTED***"
    assert mapped["note"] == "***REDACTED***"


def test_risk_scoring_and_evaluation() -> None:
    score, reasons = score_write_action("tests/unit/test_x.py", profile="full-access", dry_run=False)
    risk = evaluate_risk(score, reasons, auto_allow_max=10, approval_required_min=20, hard_block_min=50)
    assert risk.score >= 20
    assert risk.approval_required is True
