from dataclasses import dataclass
from datetime import datetime, timedelta
from pathlib import Path
from typing import Any

import yaml
from pydantic import BaseModel, Field


class EnvironmentRequest(BaseModel):
    user_id: str
    repo: str
    provider: str = "coder"
    ttl_hours: int = Field(default=8, ge=1)


@dataclass(frozen=True)
class EnvironmentState:
    environment_id: str
    request: EnvironmentRequest
    created_at: datetime

    def expires_at(self) -> datetime:
        return self.created_at + timedelta(hours=self.request.ttl_hours)

    def is_expired(self, now: datetime) -> bool:
        return now >= self.expires_at()


def load_workflow_config(path: Path) -> dict[str, Any]:
    payload: Any = yaml.safe_load(path.read_text())
    if not isinstance(payload, dict):
        raise ValueError(f"Workflow config must be a mapping: {path}")
    return payload
