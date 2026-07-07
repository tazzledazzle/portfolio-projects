from __future__ import annotations

from datetime import datetime
from datetime import timezone

from ai_image_video_generator.models import ImageVariant
from ai_image_video_generator.models import VideoClip
from ai_image_video_generator.models import VideoRequest


class VideoGenerationPipeline:
    def __init__(self, comfy_client) -> None:
        self.comfy_client = comfy_client

    def generate_clip(self, variant: ImageVariant, request: VideoRequest) -> VideoClip:
        workflow = {
            "type": "video_generation",
            "input_image_path": variant.asset_path,
            "duration_seconds": request.duration_seconds,
            "fps": request.fps,
            "seed": variant.seed + request.duration_seconds + request.fps,
        }
        result = self.comfy_client.run_workflow(workflow)
        return VideoClip(
            id=f"clip-{variant.id}",
            source_variant_id=variant.id,
            asset_path=result["asset_path"],
            seed=int(result.get("seed", workflow["seed"])),
            created_at=result.get("created_at", datetime.now(timezone.utc)),
            workflow_name=result.get("workflow_name", "video_generation"),
        )
