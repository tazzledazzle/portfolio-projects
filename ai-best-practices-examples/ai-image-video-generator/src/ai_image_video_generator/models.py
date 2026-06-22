from datetime import datetime

from pydantic import BaseModel
from pydantic import Field


class ProductBrief(BaseModel):
    product_name: str = Field(min_length=1)
    prompt: str = Field(min_length=1)
    style_notes: str = Field(default="")
    scene_description: str = Field(default="")
    variant_count: int = Field(default=3, ge=1, le=8)


class ImageQualityProfile(BaseModel):
    name: str
    sampler: str
    steps: int = Field(ge=10, le=80)
    cfg_scale: float = Field(ge=1.0, le=12.0)
    width: int = Field(ge=512, le=2048)
    height: int = Field(ge=512, le=2048)
    scheduler: str = "karras"


class ImageVariant(BaseModel):
    id: str
    prompt: str
    asset_path: str
    seed: int
    created_at: datetime
    model_name: str
    workflow_name: str


class VideoRequest(BaseModel):
    image_variant_id: str = Field(min_length=1)
    duration_seconds: int = Field(ge=1, le=8)
    fps: int = Field(ge=4, le=24)


class VideoClip(BaseModel):
    id: str
    source_variant_id: str
    asset_path: str
    seed: int
    created_at: datetime
    workflow_name: str


class GenerationRun(BaseModel):
    brief: ProductBrief
    image_variants: list[ImageVariant] = Field(default_factory=list)
    videos: list[VideoClip] = Field(default_factory=list)
