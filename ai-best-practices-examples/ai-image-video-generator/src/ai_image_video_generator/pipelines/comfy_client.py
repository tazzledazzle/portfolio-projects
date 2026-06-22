from __future__ import annotations

from dataclasses import dataclass
from typing import Any

import requests


class ComfyClientError(RuntimeError):
    """Raised when a ComfyUI API call fails or returns malformed data."""


@dataclass
class RequestsTransport:
    base_url: str
    timeout_seconds: int = 180

    def post_json(self, path: str, payload: dict[str, Any]) -> dict[str, Any]:
        response = requests.post(
            f"{self.base_url.rstrip('/')}{path}",
            json=payload,
            timeout=self.timeout_seconds,
        )
        response.raise_for_status()
        return response.json()


class ComfyClient:
    def __init__(self, base_url: str, transport: Any | None = None) -> None:
        self.base_url = base_url
        self.transport = transport if transport is not None else RequestsTransport(base_url)

    def run_workflow(self, workflow: dict[str, Any]) -> dict[str, Any]:
        result = self.transport.post_json("/api/generate", workflow)
        if "asset_path" not in result:
            raise ComfyClientError("Comfy response missing required field: asset_path")
        return result
