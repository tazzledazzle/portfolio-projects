import gradio as gr

from ai_image_video_generator.app import build_demo
from ai_image_video_generator.app import build_product_brief


def test_build_product_brief_uses_defaults() -> None:
    brief = build_product_brief(
        product_name="Sneaker",
        prompt="clean product hero",
        style_notes="minimal",
        scene_description="studio floor",
        variant_count=0,
    )
    assert brief.variant_count == 3
    assert brief.product_name == "Sneaker"


def test_build_demo_returns_blocks() -> None:
    demo = build_demo()
    assert isinstance(demo, gr.Blocks)
