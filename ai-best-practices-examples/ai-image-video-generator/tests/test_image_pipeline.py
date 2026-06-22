from datetime import datetime
from datetime import timezone

from ai_image_video_generator.models import ProductBrief
from ai_image_video_generator.pipelines.image_generation import ImageGenerationPipeline
from ai_image_video_generator.pipelines.image_generation import build_prompt


class FakeComfyClient:
    def __init__(self) -> None:
        self.calls: list[dict] = []

    def run_workflow(self, workflow: dict) -> dict:
        self.calls.append(workflow)
        idx = len(self.calls)
        return {
            "asset_path": f"outputs/images/variant-{idx}.png",
            "seed": workflow["seed"],
            "model_name": "sdxl-base",
            "workflow_name": "image_generation",
            "created_at": datetime.now(timezone.utc),
        }


def test_build_prompt_includes_product_style_scene() -> None:
    brief = ProductBrief(
        product_name="Ceramic Mug",
        prompt="premium studio product photo",
        style_notes="soft pastel palette",
        scene_description="on a wood shelf with plants",
        variant_count=3,
    )
    prompt = build_prompt(brief)
    assert "Ceramic Mug" in prompt
    assert "soft pastel palette" in prompt
    assert "wood shelf" in prompt


def test_image_generation_pipeline_creates_requested_variants() -> None:
    client = FakeComfyClient()
    pipeline = ImageGenerationPipeline(comfy_client=client)
    brief = ProductBrief(
        product_name="Headphones",
        prompt="ecommerce product shot",
        style_notes="high contrast",
        scene_description="black background",
        variant_count=2,
    )

    variants = pipeline.generate_variants(
        brief=brief,
        brand_reference_paths=["refs/brand-a.png"],
    )

    assert len(variants) == 2
    assert variants[0].asset_path.endswith("variant-1.png")
    assert variants[1].asset_path.endswith("variant-2.png")
    assert client.calls[0]["conditioning"]["controlnet_refs"] == ["refs/brand-a.png"]


def test_image_generation_pipeline_applies_quality_profile_parameters() -> None:
    client = FakeComfyClient()
    pipeline = ImageGenerationPipeline(comfy_client=client)
    brief = ProductBrief(
        product_name="Watch",
        prompt="photorealistic watch product shot",
        style_notes="luxury catalog",
        scene_description="soft gradient background",
        variant_count=1,
    )

    pipeline.generate_variants(
        brief=brief,
        brand_reference_paths=None,
        quality_profile_name="photo_studio_v1",
    )

    call = client.calls[0]
    assert call["sampler"] == "dpmpp_2m_sde"
    assert call["steps"] >= 30
    assert call["cfg_scale"] >= 4.5
    assert call["width"] >= 1024
