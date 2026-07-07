from __future__ import annotations

import json
from pathlib import Path

import gradio as gr

from ai_image_video_generator.config import QUALITY_PROFILES
from ai_image_video_generator.config import load_config
from ai_image_video_generator.export.review_gallery import build_review_gallery
from ai_image_video_generator.models import ImageVariant
from ai_image_video_generator.models import ProductBrief
from ai_image_video_generator.models import VideoClip
from ai_image_video_generator.models import VideoRequest
from ai_image_video_generator.pipelines.client_factory import active_backend_name
from ai_image_video_generator.pipelines.client_factory import create_generation_client
from ai_image_video_generator.pipelines.comfy_client import ComfyClientError
from ai_image_video_generator.pipelines.image_generation import ImageGenerationPipeline
from ai_image_video_generator.pipelines.video_generation import VideoGenerationPipeline
from ai_image_video_generator.safety.provenance import build_manifest


def build_product_brief(
    product_name: str,
    prompt: str,
    style_notes: str,
    scene_description: str,
    variant_count: int,
) -> ProductBrief:
    safe_variant_count = variant_count if variant_count >= 1 else 3
    return ProductBrief(
        product_name=product_name,
        prompt=prompt,
        style_notes=style_notes,
        scene_description=scene_description,
        variant_count=safe_variant_count,
    )


def _variant_choices(state_json: str) -> gr.Dropdown:
    state = json.loads(state_json or "{}")
    choices = [item["id"] for item in state.get("variants", [])]
    value = choices[0] if choices else None
    return gr.Dropdown(choices=choices, value=value)


def _generate_images(
    product_name: str,
    prompt: str,
    style_notes: str,
    scene_description: str,
    variant_count: int,
    brand_refs: list[str] | None,
    quality_profile_name: str,
    state_json: str,
) -> tuple[str, list[str], str, str, gr.Dropdown, str]:
    config = load_config()
    backend = active_backend_name(config)
    try:
        client = create_generation_client(config)
        image_pipeline = ImageGenerationPipeline(comfy_client=client)
        brief = build_product_brief(
            product_name=product_name,
            prompt=prompt,
            style_notes=style_notes,
            scene_description=scene_description,
            variant_count=variant_count,
        )
        variants = image_pipeline.generate_variants(
            brief=brief,
            brand_reference_paths=brand_refs,
            quality_profile_name=quality_profile_name,
        )
        state = {
            "brief": brief.model_dump(),
            "variants": [v.model_dump(mode="json") for v in variants],
            "videos": [],
        }
        paths = [v.asset_path for v in variants]
        status = f"Generated {len(variants)} image variant(s) using {backend} backend."
        serialized_state = json.dumps(state)
        return status, paths, "\n".join(paths), serialized_state, _variant_choices(serialized_state), backend
    except (ComfyClientError, ValueError, FileNotFoundError) as exc:
        return f"Image generation failed: {exc}", [], "", state_json or "{}", _variant_choices(state_json), backend


def _generate_video(
    variant_id: str,
    state_json: str,
    duration_seconds: int,
    fps: int,
) -> tuple[str, str | None, str, str]:
    config = load_config()
    backend = active_backend_name(config)
    state = json.loads(state_json or "{}")
    variants = [ImageVariant.model_validate(v) for v in state.get("variants", [])]
    variant = next((item for item in variants if item.id == variant_id), None)
    if variant is None:
        return "Variant not found in state. Generate images first.", None, state_json, backend

    try:
        client = create_generation_client(config)
        video_pipeline = VideoGenerationPipeline(comfy_client=client)
        request = VideoRequest(image_variant_id=variant_id, duration_seconds=duration_seconds, fps=fps)
        clip = video_pipeline.generate_clip(variant=variant, request=request)
        videos = state.get("videos", [])
        videos.append(clip.model_dump(mode="json"))
        state["videos"] = videos
        updated_state = json.dumps(state)
        status = f"Generated video clip for {variant_id} using {backend} backend."
        return status, clip.asset_path, updated_state, backend
    except (ComfyClientError, ValueError, FileNotFoundError) as exc:
        return f"Video generation failed: {exc}", None, state_json, backend


def _export_gallery(state_json: str) -> str:
    state = json.loads(state_json or "{}")
    output_dir = Path(load_config().output_dir)
    output_dir.mkdir(parents=True, exist_ok=True)

    image_paths = [item["asset_path"] for item in state.get("variants", [])]
    video_paths = [item["asset_path"] for item in state.get("videos", [])]
    html_path = build_review_gallery(output_dir=output_dir, image_paths=image_paths, video_paths=video_paths)

    manifest = build_manifest(
        prompt_lineage=state.get("brief", {}),
        image_variants=[ImageVariant.model_validate(v) for v in state.get("variants", [])],
        video_clips=[VideoClip.model_validate(v) for v in state.get("videos", [])],
    )
    manifest_path = output_dir / "manifest.json"
    manifest_path.write_text(json.dumps(manifest, indent=2), encoding="utf-8")
    return f"Exported: {html_path} and {manifest_path}"


def build_demo() -> gr.Blocks:
    config = load_config()
    backend = active_backend_name(config)
    with gr.Blocks(title="AI Image & Video Generator") as demo:
        gr.Markdown("## AI Image & Video Generator MVP")
        gr.Markdown(
            f"Active backend: **{backend}**. "
            "Set `AIVG_BACKEND=local|comfyui|auto` to control generation mode."
        )
        with gr.Row():
            product_name = gr.Textbox(label="Product name", value="Ceramic mug")
            prompt = gr.Textbox(label="Prompt", value="Premium studio product photo")
        with gr.Row():
            style_notes = gr.Textbox(label="Style notes", value="Soft neutral palette")
            scene_description = gr.Textbox(label="Scene description", value="On marble countertop")
        variant_count = gr.Slider(label="Variant count", minimum=1, maximum=6, step=1, value=3)
        quality_profile_name = gr.Dropdown(
            label="Quality profile",
            choices=list(QUALITY_PROFILES.keys()),
            value=config.default_quality_profile,
        )
        brand_refs = gr.File(label="Brand references", file_count="multiple", type="filepath")
        generate_images_button = gr.Button("Generate image variants")
        image_status = gr.Textbox(label="Image generation status")
        image_gallery = gr.Gallery(label="Generated image variants", columns=3, height="auto")
        image_paths = gr.Textbox(label="Generated image paths", lines=3)

        state = gr.Textbox(label="Run state (internal)", visible=False)
        backend_label = gr.Textbox(label="Backend (internal)", visible=False, value=backend)
        variant_id = gr.Dropdown(label="Variant ID for video", choices=[], value=None)
        duration_seconds = gr.Slider(label="Video duration seconds", minimum=1, maximum=8, step=1, value=4)
        fps = gr.Slider(label="Video FPS", minimum=4, maximum=24, step=1, value=8)
        generate_video_button = gr.Button("Generate video")
        video_status = gr.Textbox(label="Video generation status")
        video_output = gr.File(label="Generated video clip")

        export_button = gr.Button("Export review gallery")
        export_status = gr.Textbox(label="Export status")

        generate_images_button.click(
            fn=_generate_images,
            inputs=[
                product_name,
                prompt,
                style_notes,
                scene_description,
                variant_count,
                brand_refs,
                quality_profile_name,
                state,
            ],
            outputs=[image_status, image_gallery, image_paths, state, variant_id, backend_label],
        )
        generate_video_button.click(
            fn=_generate_video,
            inputs=[variant_id, state, duration_seconds, fps],
            outputs=[video_status, video_output, state, backend_label],
        )
        export_button.click(fn=_export_gallery, inputs=[state], outputs=[export_status])
    return demo


def main() -> None:
    demo = build_demo()
    demo.launch()


if __name__ == "__main__":
    main()
