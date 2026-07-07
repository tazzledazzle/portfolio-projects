from __future__ import annotations

from dataclasses import dataclass
from typing import Any

import requests


class ComfyClientError(RuntimeError):
    """Raised when a ComfyUI API call fails or returns malformed data."""


def is_comfyui_available(base_url: str, timeout_seconds: float = 2.0) -> bool:
    try:
        response = requests.get(f"{base_url.rstrip('/')}/system_stats", timeout=timeout_seconds)
        return response.ok
    except (requests.RequestException, OSError):
        return False


@dataclass
class RequestsTransport:
    base_url: str
    timeout_seconds: int = 180

    def post_json(self, path: str, payload: dict[str, Any]) -> dict[str, Any]:
        url = f"{self.base_url.rstrip('/')}{path}"
        try:
            response = requests.post(url, json=payload, timeout=self.timeout_seconds)
            response.raise_for_status()
        except requests.ConnectionError as exc:
            raise ComfyClientError(
                f"Could not reach ComfyUI at {self.base_url}. "
                "Start ComfyUI or set AIVG_BACKEND=local for offline generation."
            ) from exc
        except requests.HTTPError as exc:
            raise ComfyClientError(
                f"ComfyUI request failed ({response.status_code}) for {url}. "
                "Ensure your ComfyUI server exposes /api/generate."
            ) from exc
        except requests.RequestException as exc:
            raise ComfyClientError(f"ComfyUI request failed for {url}: {exc}") from exc
        return response.json()


class ComfyClient:
    def __init__(self, base_url: str, transport: Any | None = None) -> None:
        self.base_url = base_url
        self.transport = transport if transport is not None else RequestsTransport(base_url)

    def run_workflow(self, workflow: dict[str, Any]) -> dict[str, Any]:
        try:
            result = self.transport.post_json("/api/generate", workflow)
        except requests.ConnectionError as exc:
            raise ComfyClientError(
                f"Could not reach ComfyUI at {self.base_url}. "
                "Start ComfyUI or set AIVG_BACKEND=local for offline generation."
            ) from exc
        except requests.HTTPError as exc:
            raise ComfyClientError(
                f"ComfyUI request failed for {self.base_url}/api/generate. "
                "Ensure your ComfyUI server exposes /api/generate."
            ) from exc
        except requests.RequestException as exc:
            raise ComfyClientError(f"ComfyUI request failed for {self.base_url}: {exc}") from exc
        if "asset_path" not in result:
            raise ComfyClientError("Comfy response missing required field: asset_path")
        return result
