from __future__ import annotations

from pathlib import Path
from typing import Any

from PIL import Image
from PIL import ImageDraw


def apply_image_watermark(source_path: Path, output_path: Path, text: str) -> None:
    image = Image.open(source_path).convert("RGBA")
    draw = ImageDraw.Draw(image)
    x = 8
    y = max(8, image.height - 20)
    draw.rectangle((x - 4, y - 2, x + 120, y + 14), fill=(0, 0, 0, 140))
    draw.text((x, y), text, fill=(255, 255, 255, 230))
    image.convert("RGB").save(output_path)


def build_manifest(
    prompt_lineage: dict[str, Any],
    image_variants: list[Any],
    video_clips: list[Any],
) -> dict[str, Any]:
    return {
        "prompt_lineage": prompt_lineage,
        "images": [
            {
                "id": variant.id,
                "asset_path": variant.asset_path,
                "seed": variant.seed,
                "model_name": getattr(variant, "model_name", ""),
                "workflow_name": variant.workflow_name,
                "created_at": variant.created_at.isoformat(),
            }
            for variant in image_variants
        ],
        "videos": [
            {
                "id": clip.id,
                "source_variant_id": clip.source_variant_id,
                "asset_path": clip.asset_path,
                "seed": clip.seed,
                "workflow_name": clip.workflow_name,
                "created_at": clip.created_at.isoformat(),
            }
            for clip in video_clips
        ],
    }
