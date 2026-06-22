import os
import tomllib
from dataclasses import dataclass, field
from pathlib import Path


@dataclass(frozen=True)
class AssistantPolicy:
    auto_allow_max: int = 30
    approval_required_min: int = 31
    hard_block_min: int = 80
    redaction_enabled: bool = True
    redaction_patterns: list[str] = field(
        default_factory=lambda: [r"sk-[A-Za-z0-9]{10,}", r"ghp_[A-Za-z0-9]{20,}"]
    )


def load_policy(policy_file: str | None) -> tuple[AssistantPolicy, str]:
    candidates: list[Path] = []
    if policy_file:
        candidates.append(Path(policy_file).resolve())
    env_file = os.getenv("AI_CODE_ASSISTANT_POLICY_FILE")
    if env_file:
        candidates.append(Path(env_file).resolve())
    candidates.append(Path.cwd() / "assistant-policy.toml")

    for path in candidates:
        if path.exists():
            data = tomllib.loads(path.read_text(encoding="utf-8"))
            risk = data.get("risk", {})
            redaction = data.get("redaction", {})
            patterns = redaction.get("patterns")
            policy = AssistantPolicy(
                auto_allow_max=int(risk.get("auto_allow_max", 30)),
                approval_required_min=int(risk.get("approval_required_min", 31)),
                hard_block_min=int(risk.get("hard_block_min", 80)),
                redaction_enabled=bool(redaction.get("enabled", True)),
                redaction_patterns=list(patterns) if isinstance(patterns, list) else AssistantPolicy().redaction_patterns,
            )
            return policy, str(path)

    return AssistantPolicy(), "defaults"
