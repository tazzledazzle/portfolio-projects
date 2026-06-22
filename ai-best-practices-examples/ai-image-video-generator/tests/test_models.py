from ai_image_video_generator.models import ProductBrief
from ai_image_video_generator.models import VideoRequest


def test_product_brief_rejects_empty_product_name() -> None:
    try:
        ProductBrief(
            product_name="",
            prompt="clean studio photo",
            style_notes="minimal",
            scene_description="on white table",
            variant_count=3,
        )
    except ValueError as exc:
        assert "product_name" in str(exc)
    else:
        raise AssertionError("Expected empty product_name to be rejected")


def test_video_request_rejects_invalid_duration() -> None:
    try:
        VideoRequest(
            image_variant_id="variant-a",
            duration_seconds=0,
            fps=8,
        )
    except ValueError as exc:
        assert "duration_seconds" in str(exc)
    else:
        raise AssertionError("Expected invalid duration to be rejected")
