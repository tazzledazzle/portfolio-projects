from datetime import datetime, timedelta
from pathlib import Path

from activities.teardown.ttl_teardown import build_environment_state, should_teardown, teardown_deadline
from domain.models import EnvironmentRequest, load_workflow_config
from workflows.dev_environment_workflow import run_dev_environment_workflow


def test_run_dev_environment_workflow_schedules_request() -> None:
    result = run_dev_environment_workflow(
        {"user_id": "dev1", "repo": "portfolio-projects", "provider": "coder", "ttl_hours": 4}
    )
    assert result["status"] == "scheduled"
    assert result["ttl_hours"] == 4


def test_should_teardown_after_ttl() -> None:
    created = datetime(2026, 1, 1, 12, 0, 0)
    request = EnvironmentRequest(user_id="dev1", repo="demo", ttl_hours=2)
    state = build_environment_state("env-1", request, created_at=created)
    assert should_teardown(state, created + timedelta(hours=2)) is True
    assert should_teardown(state, created + timedelta(hours=1)) is False


def test_load_workflow_config_reads_ttl() -> None:
    root = Path(__file__).resolve().parents[2]
    config = load_workflow_config(root / "config/workflow.yaml")
    assert config["workflow"]["ttl_hours"] == 8


def test_teardown_deadline_matches_state() -> None:
    created = datetime(2026, 1, 1, 12, 0, 0)
    request = EnvironmentRequest(user_id="dev1", repo="demo", ttl_hours=8)
    state = build_environment_state("env-1", request, created_at=created)
    assert teardown_deadline(created, 8) == state.expires_at()
