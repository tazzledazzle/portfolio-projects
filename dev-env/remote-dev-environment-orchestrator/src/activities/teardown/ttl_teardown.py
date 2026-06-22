from datetime import datetime, timedelta

from domain.models import EnvironmentRequest, EnvironmentState


def should_teardown(state: EnvironmentState, now: datetime) -> bool:
    return state.is_expired(now)


def default_ttl_hours(config: dict) -> int:
    workflow = config.get("workflow", {})
    return int(workflow.get("ttl_hours", 8))


def build_environment_state(
    environment_id: str,
    request: EnvironmentRequest,
    created_at: datetime | None = None,
) -> EnvironmentState:
    return EnvironmentState(
        environment_id=environment_id,
        request=request,
        created_at=created_at or datetime.utcnow(),
    )


def teardown_deadline(created_at: datetime, ttl_hours: int) -> datetime:
    return created_at + timedelta(hours=ttl_hours)
