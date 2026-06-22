from datetime import datetime
from datetime import timezone

from ai_image_video_generator.export.review_gallery import build_review_gallery
from ai_image_video_generator.models import ImageVariant
from ai_image_video_generator.models import VideoClip
from ai_image_video_generator.safety.provenance import build_manifest


def test_build_manifest_contains_images_and_videos() -> None:
    image = ImageVariant(
        id="variant-1",
        prompt="product prompt",
        asset_path="outputs/images/variant-1.png",
        seed=101,
        created_at=datetime.now(timezone.utc),
        model_name="sdxl-base",
        workflow_name="image_generation",
    )
    video = VideoClip(
        id="clip-1",
        source_variant_id="variant-1",
        asset_path="outputs/videos/clip-1.mp4",
        seed=202,
        created_at=datetime.now(timezone.utc),
        workflow_name="video_generation",
    )

    manifest = build_manifest(
        prompt_lineage={"product_name": "Mug"},
        image_variants=[image],
        video_clips=[video],
    )
    assert manifest["prompt_lineage"]["product_name"] == "Mug"
    assert manifest["images"][0]["id"] == "variant-1"
    assert manifest["videos"][0]["id"] == "clip-1"


def test_build_review_gallery_writes_html(tmp_path) -> None:
    image_paths = ["outputs/images/variant-1.png"]
    video_paths = ["outputs/videos/clip-1.mp4"]
    html_path = build_review_gallery(
        output_dir=tmp_path,
        image_paths=image_paths,
        video_paths=video_paths,
    )
    html = html_path.read_text(encoding="utf-8")
    assert "variant-1.png" in html
    assert "clip-1.mp4" in html
