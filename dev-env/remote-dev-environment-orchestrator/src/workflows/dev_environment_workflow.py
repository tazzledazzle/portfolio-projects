from domain.models import EnvironmentRequest


def run_dev_environment_workflow(request_payload: dict) -> dict:
    request = EnvironmentRequest.model_validate(request_payload)
    return {
        "status": "scheduled",
        "user_id": request.user_id,
        "repo": request.repo,
        "provider": request.provider,
        "ttl_hours": request.ttl_hours,
    }
