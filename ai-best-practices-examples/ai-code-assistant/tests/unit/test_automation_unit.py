import json
from pathlib import Path

from ai_code_assistant.automation import load_plan


def test_load_plan_reads_steps(tmp_path: Path) -> None:
    plan = tmp_path / "plan.json"
    plan.write_text(
        json.dumps({"steps": [{"id": "s1", "command": "echo ok", "verify_command": "echo verified"}]}),
        encoding="utf-8",
    )
    steps = load_plan(plan)
    assert len(steps) == 1
    assert steps[0].id == "s1"
