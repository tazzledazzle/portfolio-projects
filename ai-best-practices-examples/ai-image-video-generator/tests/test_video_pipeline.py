from datetime import datetime
from datetime import timezone

from ai_image_video_generator.models import ImageVariant
from ai_image_video_generator.models import VideoRequest
from ai_image_video_generator.pipelines.video_generation import VideoGenerationPipeline


class FakeComfyVideoClient:
    def __init__(self) -> None:
        self.calls: list[dict] = []

    def run_workflow(self, workflow: dict) -> dict:
        self.calls.append(workflow)
        return {
            "asset_path": "outputs/videos/clip-1.mp4",
            "seed": workflow["seed"],
            "workflow_name": "video_generation",
            "created_at": datetime.now(timezone.utc),
        }


def test_video_pipeline_builds_video_workflow_from_variant() -> None:
    variant = ImageVariant(
        id="variant-1",
        prompt="test prompt",
        asset_path="outputs/images/variant-1.png",
        seed=123,
        created_at=datetime.now(timezone.utc),
        model_name="sdxl-base",
        workflow_name="image_generation",
    )
    request = VideoRequest(image_variant_id="variant-1", duration_seconds=4, fps=8)
    pipeline = VideoGenerationPipeline(comfy_client=FakeComfyVideoClient())

    clip = pipeline.generate_clip(variant=variant, request=request)

    assert clip.asset_path.endswith(".mp4")
    assert clip.source_variant_id == "variant-1"
    assert clip.workflow_name == "video_generation"
