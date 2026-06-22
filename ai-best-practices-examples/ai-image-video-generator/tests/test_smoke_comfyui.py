import os

import pytest

from ai_image_video_generator.app import build_product_brief
from ai_image_video_generator.pipelines.comfy_client import ComfyClient
from ai_image_video_generator.pipelines.image_generation import ImageGenerationPipeline


@pytest.mark.skipif(
    os.getenv("AIVG_RUN_SMOKE") != "1",
    reason="Set AIVG_RUN_SMOKE=1 to run live ComfyUI smoke test.",
)
def test_live_comfyui_image_smoke() -> None:
    client = ComfyClient(base_url=os.getenv("AIVG_COMFYUI_BASE_URL", "http://127.0.0.1:8188"))
    pipeline = ImageGenerationPipeline(comfy_client=client)
    brief = build_product_brief(
        product_name="Smoke Test Product",
        prompt="simple catalog shot",
        style_notes="neutral",
        scene_description="white background",
        variant_count=1,
    )
    variants = pipeline.generate_variants(brief=brief, brand_reference_paths=None)
    assert len(variants) == 1
