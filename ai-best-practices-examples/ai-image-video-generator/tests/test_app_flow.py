import json

from ai_image_video_generator.app import _export_gallery
from ai_image_video_generator.app import _generate_images
from ai_image_video_generator.app import _generate_video


def test_generate_images_local_backend(monkeypatch, tmp_path) -> None:
    monkeypatch.setenv("AIVG_BACKEND", "local")
    monkeypatch.setenv("AIVG_OUTPUT_DIR", str(tmp_path))

    status, paths, path_text, state_json, variant_dropdown, backend = _generate_images(
        product_name="Ceramic mug",
        prompt="Premium studio product photo",
        style_notes="Soft neutral palette",
        scene_description="On marble countertop",
        variant_count=2,
        brand_refs=None,
        quality_profile_name="photo_studio_v1",
        state_json="{}",
    )

    assert backend == "local"
    assert "Generated 2 image variant(s)" in status
    assert len(paths) == 2
    assert path_text.count("\n") == 1
    state = json.loads(state_json)
    assert len(state["variants"]) == 2
    choice_values = [choice[0] if isinstance(choice, tuple) else choice for choice in variant_dropdown.choices]
    assert choice_values == ["variant-1", "variant-2"]


def test_generate_video_local_backend(monkeypatch, tmp_path) -> None:
    monkeypatch.setenv("AIVG_BACKEND", "local")
    monkeypatch.setenv("AIVG_OUTPUT_DIR", str(tmp_path))

    _, _, _, state_json, _, _ = _generate_images(
        product_name="Ceramic mug",
        prompt="Premium studio product photo",
        style_notes="",
        scene_description="",
        variant_count=1,
        brand_refs=None,
        quality_profile_name="photo_studio_v1",
        state_json="{}",
    )

    status, video_path, updated_state, backend = _generate_video(
        variant_id="variant-1",
        state_json=state_json,
        duration_seconds=2,
        fps=8,
    )

    assert backend == "local"
    assert "Generated video clip" in status
    assert video_path is not None
    assert updated_state
    export_status = _export_gallery(updated_state)
    assert "review_gallery.html" in export_status
    assert "manifest.json" in export_status
