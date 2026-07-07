from __future__ import annotations

from datetime import datetime
from datetime import timezone

from ai_image_video_generator.config import get_quality_profile
from ai_image_video_generator.models import ImageVariant
from ai_image_video_generator.models import ProductBrief
from ai_image_video_generator.pipelines.style_conditioning import build_conditioning_payload


def build_prompt(brief: ProductBrief) -> str:
    segments = [
        brief.prompt.strip(),
        f"Product: {brief.product_name.strip()}",
    ]
    if brief.style_notes.strip():
        segments.append(f"Style: {brief.style_notes.strip()}")
    if brief.scene_description.strip():
        segments.append(f"Scene: {brief.scene_description.strip()}")
    return ". ".join(segments)


def build_negative_prompt() -> str:
    return (
        "blurry, distorted, low quality, cgi, plastic texture, oversmoothed, "
        "overexposed, underexposed, jpeg artifacts"
    )


class ImageGenerationPipeline:
    def __init__(self, comfy_client) -> None:
        self.comfy_client = comfy_client

    def generate_variants(
        self,
        brief: ProductBrief,
        brand_reference_paths: list[str] | None = None,
        quality_profile_name: str = "photo_studio_v1",
    ) -> list[ImageVariant]:
        prompt = build_prompt(brief)
        conditioning = build_conditioning_payload(brand_reference_paths)
        profile = get_quality_profile(quality_profile_name)
        variants: list[ImageVariant] = []

        for idx in range(brief.variant_count):
            workflow = {
                "type": "image_generation",
                "prompt": prompt,
                "negative_prompt": build_negative_prompt(),
                "seed": 1000 + idx,
                "conditioning": conditioning,
                "sampler": profile.sampler,
                "steps": profile.steps,
                "cfg_scale": profile.cfg_scale,
                "width": profile.width,
                "height": profile.height,
                "scheduler": profile.scheduler,
                "quality_profile": profile.name,
            }
            result = self.comfy_client.run_workflow(workflow)
            variants.append(
                ImageVariant(
                    id=f"variant-{idx + 1}",
                    prompt=prompt,
                    asset_path=result["asset_path"],
                    seed=int(result.get("seed", workflow["seed"])),
                    created_at=result.get("created_at", datetime.now(timezone.utc)),
                    model_name=result.get("model_name", "unknown-model"),
                    workflow_name=result.get("workflow_name", "image_generation"),
                )
            )
        return variants
