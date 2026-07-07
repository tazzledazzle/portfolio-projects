# AI Image & Video Generator

Python + Gradio MVP for product image/video generation with ComfyUI orchestration.

## Features

- Generate multiple product image variants from a text brief.
- Apply brand-style conditioning hooks (ControlNet/IP-Adapter references).
- Generate short video clips from selected image variants.
- Export a review bundle with `review_gallery.html` and `manifest.json`.
- Add baseline provenance and image watermarking utilities.

## Setup

```bash
cd ai-image-video-generator
python3 -m pip install -e ".[dev]"
```

## Run

### Start the Gradio app

```bash
python3 -m ai_image_video_generator.app
```

### Automated local launch sanity check

```bash
./scripts/launch_sanity_check.sh
```

### Run tests

```bash
python3 -m pytest -q
```

### Optional live ComfyUI smoke test

Requires a reachable ComfyUI endpoint implementing `/api/generate`:

```bash
AIVG_RUN_SMOKE=1 python3 -m pytest -q tests/test_smoke_comfyui.py
```

## Configuration

Configure the app with environment variables:

- `AIVG_COMFYUI_BASE_URL` (default `http://127.0.0.1:8188`)
- `AIVG_IMAGE_WORKFLOW_PATH` (default `workflows/image_generation.json`)
- `AIVG_VIDEO_WORKFLOW_PATH` (default `workflows/video_generation.json`)
- `AIVG_OUTPUT_DIR` (default `outputs`)
- `AIVG_DEFAULT_VARIANT_COUNT` (default `3`)
- `AIVG_DEFAULT_QUALITY_PROFILE` (default `photo_studio_v1`)

Example:

```bash
export AIVG_COMFYUI_BASE_URL="http://127.0.0.1:8188"
export AIVG_OUTPUT_DIR="outputs"
export AIVG_DEFAULT_QUALITY_PROFILE="photo_studio_v1"
python3 -m ai_image_video_generator.app
```

## Best Practices

- Use `photo_studio_v1` as your default profile for realistic catalog-style product images.
- Start with `variant_count=3`, then increase only when you need wider exploration.
- Provide 1-3 strong brand reference images to improve style consistency.
- Keep prompts concrete (product, material, lighting, scene) and avoid ambiguous adjectives.
- Validate every local change with `python3 -m pytest -q` before relying on results.
- Keep workflows pinned and versioned in `workflows/` to avoid ComfyUI graph drift.
- Use export artifacts (`review_gallery.html` and `manifest.json`) for traceability and review.

## Examples

### 1) Basic run

```bash
python3 -m ai_image_video_generator.app
```

In the UI:
- Set product name: `Ceramic mug`
- Prompt: `Premium studio product photo`
- Quality profile: `photo_studio_v1`
- Variant count: `3`

### 2) Lifestyle composition

Use:
- Quality profile: `photo_lifestyle_v1`
- Scene description: `On a kitchen counter with soft morning light`
- Style notes: `Editorial, natural shadows, realistic textures`

### 3) Macro detail pass

Use:
- Quality profile: `photo_macro_v1`
- Prompt: `Ultra-detailed macro product shot of a stainless steel watch`
- Style notes: `High micro-contrast, realistic reflections, no CGI look`

### 4) CI-friendly launch check

```bash
./scripts/launch_sanity_check.sh
```

## FAQ

### Which quality profile should I start with?
Start with `photo_studio_v1`. It is tuned for balanced photorealism in product shots.

### Why are images not generating?
Check that:
- ComfyUI is running.
- `AIVG_COMFYUI_BASE_URL` points to the correct host/port.
- Your ComfyUI endpoint supports the expected `/api/generate` contract.

### What if outputs look synthetic?
- Use stronger product/material language in prompt text.
- Add cleaner reference images.
- Try `photo_macro_v1` for detail-heavy shots.
- Increase steps carefully in your workflow-level tuning.

### Where are generated files and exports?
By default, outputs go under `outputs/`. Export generates `review_gallery.html` and `manifest.json`.

### Can I use my own ComfyUI workflows?
Yes. Keep workflow files under `workflows/` and point paths via:
- `AIVG_IMAGE_WORKFLOW_PATH`
- `AIVG_VIDEO_WORKFLOW_PATH`

## Notes

- Workflow JSON files in `workflows/` are placeholders to pin names and schema location.
- Replace them with concrete ComfyUI graph exports for production runs.
