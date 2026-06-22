from PIL import Image

from ai_image_video_generator.safety.provenance import apply_image_watermark


def test_apply_image_watermark_writes_file(tmp_path) -> None:
    source = tmp_path / "source.png"
    out = tmp_path / "watermarked.png"
    Image.new("RGB", (128, 128), color="white").save(source)

    apply_image_watermark(
        source_path=source,
        output_path=out,
        text="AI Generated",
    )

    assert out.exists()
