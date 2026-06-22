from pathlib import Path
import sys

sys.path.append(str(Path(__file__).resolve().parents[1] / "src"))

from main import rollout_step


def test_rollout_rollback_when_slo_breaches() -> None:
    decision = rollout_step(40, error_rate=0.05, p99_latency_ms=800)
    assert decision["action"] == "rollback"


def test_rollout_advances_when_healthy() -> None:
    decision = rollout_step(40, error_rate=0.01, p99_latency_ms=300)
    assert decision["next_weight"] == 60
