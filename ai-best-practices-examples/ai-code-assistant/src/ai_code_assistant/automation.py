import json
import subprocess
from concurrent.futures import ThreadPoolExecutor
from dataclasses import dataclass
from pathlib import Path


@dataclass(frozen=True)
class PlanStep:
    id: str
    command: str
    verify_command: str
    high_risk: bool = False


def load_plan(path: Path) -> list[PlanStep]:
    payload = json.loads(path.read_text(encoding="utf-8"))
    return [
        PlanStep(
            id=step["id"],
            command=step["command"],
            verify_command=step.get("verify_command", "true"),
            high_risk=bool(step.get("high_risk", False)),
        )
        for step in payload.get("steps", [])
    ]


def run_step(step: PlanStep, checkpoint_dir: Path) -> dict:
    checkpoint_dir.mkdir(parents=True, exist_ok=True)
    state_file = checkpoint_dir / f"{step.id}.json"
    state_file.write_text(json.dumps({"status": "running"}), encoding="utf-8")
    subprocess.run(step.command, shell=True, check=True)
    subprocess.run(step.verify_command, shell=True, check=True)
    state_file.write_text(json.dumps({"status": "verified"}), encoding="utf-8")
    return {"id": step.id, "status": "verified"}


def run_steps(steps: list[PlanStep], checkpoint_dir: Path, parallel: bool) -> list[dict]:
    if parallel:
        with ThreadPoolExecutor(max_workers=4) as executor:
            return list(executor.map(lambda s: run_step(s, checkpoint_dir), steps))
    return [run_step(step, checkpoint_dir) for step in steps]
