from __future__ import annotations

import random
import textwrap
from datetime import datetime
from datetime import timezone
from pathlib import Path
from typing import Any

from PIL import Image
from PIL import ImageDraw

from ai_image_video_generator.safety.provenance import apply_image_watermark


class LocalGenerationClient:
    """Generates placeholder images and short clips without an external ComfyUI server."""

    def __init__(self, output_dir: str | Path) -> None:
        self.output_dir = Path(output_dir)
        self.images_dir = self.output_dir / "images"
        self.videos_dir = self.output_dir / "videos"
        self.images_dir.mkdir(parents=True, exist_ok=True)
        self.videos_dir.mkdir(parents=True, exist_ok=True)

    def run_workflow(self, workflow: dict[str, Any]) -> dict[str, Any]:
        workflow_type = workflow.get("type")
        if workflow_type == "image_generation":
            return self._generate_image(workflow)
        if workflow_type == "video_generation":
            return self._generate_video(workflow)
        raise ValueError(f"Unsupported local workflow type: {workflow_type!r}")

    def _generate_image(self, workflow: dict[str, Any]) -> dict[str, Any]:
        seed = int(workflow.get("seed", 42))
        width = int(workflow.get("width", 1024))
        height = int(workflow.get("height", 1024))
        prompt = str(workflow.get("prompt", "Product image"))
        quality_profile = str(workflow.get("quality_profile", "photo_studio_v1"))

        rng = random.Random(seed)
        base_color = (rng.randint(48, 180), rng.randint(48, 180), rng.randint(48, 180))
        accent = tuple(min(255, channel + 40) for channel in base_color)

        image = Image.new("RGB", (width, height), base_color)
        draw = ImageDraw.Draw(image)
        margin = max(32, width // 20)
        draw.rectangle(
            (margin, margin, width - margin, height - margin),
            outline=accent,
            width=max(2, width // 256),
        )

        header = f"{quality_profile} | seed {seed}"
        wrapped = textwrap.fill(prompt, width=42)
        y = margin + 24
        draw.text((margin + 12, y), header, fill=(245, 245, 245))
        y += 28
        for line in wrapped.splitlines():
            draw.text((margin + 12, y), line, fill=(250, 250, 250))
            y += 22

        refs = workflow.get("conditioning", {}).get("controlnet_refs", [])
        if refs:
            draw.text((margin + 12, height - margin - 28), f"Brand refs: {len(refs)}", fill=(230, 230, 230))

        raw_path = self.images_dir / f"variant-{seed}.png"
        image.save(raw_path)

        final_path = self.images_dir / f"variant-{seed}-watermarked.png"
        apply_image_watermark(source_path=raw_path, output_path=final_path, text="AI Generated")
        raw_path.unlink(missing_ok=True)

        return {
            "asset_path": str(final_path),
            "seed": seed,
            "model_name": "local-placeholder-v1",
            "workflow_name": "image_generation",
            "created_at": datetime.now(timezone.utc),
        }

    def _generate_video(self, workflow: dict[str, Any]) -> dict[str, Any]:
        seed = int(workflow.get("seed", 42))
        duration_seconds = int(workflow.get("duration_seconds", 4))
        fps = int(workflow.get("fps", 8))
        input_image_path = Path(str(workflow.get("input_image_path", "")))

        if not input_image_path.is_file():
            raise FileNotFoundError(f"Source image not found for video generation: {input_image_path}")

        base_image = Image.open(input_image_path).convert("RGB")
        frame_count = max(2, duration_seconds * fps)
        frames = self._build_motion_frames(base_image, frame_count, seed)

        output_path = self.videos_dir / f"clip-{seed}.mp4"
        written_path = self._write_video(frames, output_path, fps)

        return {
            "asset_path": str(written_path),
            "seed": seed,
            "workflow_name": "video_generation",
            "created_at": datetime.now(timezone.utc),
        }

    def _build_motion_frames(self, base_image: Image.Image, frame_count: int, seed: int) -> list[Image.Image]:
        rng = random.Random(seed)
        width, height = base_image.size
        frames: list[Image.Image] = []
        max_shift = max(8, min(width, height) // 40)

        for index in range(frame_count):
            progress = index / max(frame_count - 1, 1)
            zoom = 1.0 + progress * 0.08
            shift_x = int(rng.uniform(-max_shift, max_shift) * progress)
            shift_y = int(rng.uniform(-max_shift, max_shift) * (1.0 - progress))

            scaled_w = int(width * zoom)
            scaled_h = int(height * zoom)
            scaled = base_image.resize((scaled_w, scaled_h), Image.Resampling.LANCZOS)

            left = max(0, min((scaled_w - width) // 2 + shift_x, scaled_w - width))
            top = max(0, min((scaled_h - height) // 2 + shift_y, scaled_h - height))
            cropped = scaled.crop((left, top, left + width, top + height))
            frames.append(cropped)
        return frames

    def _write_video(self, frames: list[Image.Image], output_path: Path, fps: int) -> Path:
        try:
            import imageio.v3 as iio

            iio.imwrite(
                output_path,
                [frame.copy() for frame in frames],
                fps=fps,
                codec="libx264",
                plugin="ffmpeg",
            )
            return output_path
        except Exception:
            gif_path = output_path.with_suffix(".gif")
            duration_ms = max(1, int(1000 / fps))
            frames[0].save(
                gif_path,
                save_all=True,
                append_images=frames[1:],
                duration=duration_ms,
                loop=0,
            )
            return gif_path
