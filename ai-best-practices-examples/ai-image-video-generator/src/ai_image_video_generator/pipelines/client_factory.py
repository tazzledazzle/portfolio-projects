from __future__ import annotations

from typing import Any
from typing import Protocol

from ai_image_video_generator.config import AppConfig
from ai_image_video_generator.pipelines.comfy_client import ComfyClient
from ai_image_video_generator.pipelines.comfy_client import is_comfyui_available
from ai_image_video_generator.pipelines.local_client import LocalGenerationClient


class GenerationClient(Protocol):
    def run_workflow(self, workflow: dict[str, Any]) -> dict[str, Any]: ...


def create_generation_client(config: AppConfig) -> GenerationClient:
    backend = config.backend.lower()
    if backend == "local":
        return LocalGenerationClient(output_dir=config.output_dir)
    if backend == "comfyui":
        return ComfyClient(base_url=config.comfyui_base_url)
    if is_comfyui_available(config.comfyui_base_url):
        return ComfyClient(base_url=config.comfyui_base_url)
    return LocalGenerationClient(output_dir=config.output_dir)


def active_backend_name(config: AppConfig) -> str:
    client = create_generation_client(config)
    if isinstance(client, LocalGenerationClient):
        return "local"
    return "comfyui"
