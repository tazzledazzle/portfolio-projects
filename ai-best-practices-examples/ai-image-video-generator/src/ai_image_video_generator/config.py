from collections.abc import Mapping
import os
from typing import Literal

from pydantic import BaseModel
from pydantic import Field

from ai_image_video_generator.models import ImageQualityProfile

BackendMode = Literal["auto", "comfyui", "local"]


class AppConfig(BaseModel):
    backend: BackendMode = "auto"
    comfyui_base_url: str = "http://127.0.0.1:8188"
    image_workflow_path: str = "workflows/image_generation.json"
    video_workflow_path: str = "workflows/video_generation.json"
    output_dir: str = "outputs"
    default_variant_count: int = Field(default=3, ge=1, le=8)
    default_quality_profile: str = "photo_studio_v1"


def load_config(env: Mapping[str, str] | None = None) -> AppConfig:
    source = dict(env) if env is not None else os.environ
    backend = source.get("AIVG_BACKEND", "auto").lower()
    if backend not in {"auto", "comfyui", "local"}:
        raise ValueError(f"Unsupported AIVG_BACKEND value: {backend!r}")
    return AppConfig(
        backend=backend,  # type: ignore[arg-type]
        comfyui_base_url=source.get("AIVG_COMFYUI_BASE_URL", "http://127.0.0.1:8188"),
        image_workflow_path=source.get(
            "AIVG_IMAGE_WORKFLOW_PATH", "workflows/image_generation.json"
        ),
        video_workflow_path=source.get(
            "AIVG_VIDEO_WORKFLOW_PATH", "workflows/video_generation.json"
        ),
        output_dir=source.get("AIVG_OUTPUT_DIR", "outputs"),
        default_variant_count=int(source.get("AIVG_DEFAULT_VARIANT_COUNT", "3")),
        default_quality_profile=source.get("AIVG_DEFAULT_QUALITY_PROFILE", "photo_studio_v1"),
    )


QUALITY_PROFILES: dict[str, ImageQualityProfile] = {
    "photo_studio_v1": ImageQualityProfile(
        name="photo_studio_v1",
        sampler="dpmpp_2m_sde",
        steps=36,
        cfg_scale=5.5,
        width=1024,
        height=1024,
        scheduler="karras",
    ),
    "photo_macro_v1": ImageQualityProfile(
        name="photo_macro_v1",
        sampler="dpmpp_2m_sde",
        steps=42,
        cfg_scale=5.0,
        width=1024,
        height=1024,
        scheduler="karras",
    ),
    "photo_lifestyle_v1": ImageQualityProfile(
        name="photo_lifestyle_v1",
        sampler="dpmpp_2m_sde",
        steps=34,
        cfg_scale=6.0,
        width=1216,
        height=832,
        scheduler="karras",
    ),
}


def get_quality_profile(name: str) -> ImageQualityProfile:
    if name not in QUALITY_PROFILES:
        raise ValueError(f"Unknown quality profile: {name}")
    return QUALITY_PROFILES[name]
