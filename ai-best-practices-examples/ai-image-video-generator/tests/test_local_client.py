from pathlib import Path

from ai_image_video_generator.config import load_config
from ai_image_video_generator.pipelines.client_factory import active_backend_name
from ai_image_video_generator.pipelines.client_factory import create_generation_client
from ai_image_video_generator.pipelines.comfy_client import ComfyClient
from ai_image_video_generator.pipelines.comfy_client import is_comfyui_available
from ai_image_video_generator.pipelines.local_client import LocalGenerationClient


def test_create_generation_client_uses_local_backend_when_forced() -> None:
    config = load_config({"AIVG_BACKEND": "local", "AIVG_OUTPUT_DIR": "outputs-test"})
    client = create_generation_client(config)
    assert isinstance(client, LocalGenerationClient)


def test_create_generation_client_uses_comfyui_when_forced(monkeypatch) -> None:
    monkeypatch.setattr(
        "ai_image_video_generator.pipelines.client_factory.is_comfyui_available",
        lambda _url: False,
    )
    config = load_config({"AIVG_BACKEND": "comfyui"})
    client = create_generation_client(config)
    assert isinstance(client, ComfyClient)


def test_auto_backend_falls_back_to_local_when_comfyui_unreachable(monkeypatch) -> None:
    monkeypatch.setattr(
        "ai_image_video_generator.pipelines.client_factory.is_comfyui_available",
        lambda _url: False,
    )
    config = load_config({"AIVG_BACKEND": "auto", "AIVG_OUTPUT_DIR": "outputs-test"})
    assert active_backend_name(config) == "local"


def test_local_client_generates_watermarked_image(tmp_path) -> None:
    client = LocalGenerationClient(output_dir=tmp_path)
    result = client.run_workflow(
        {
            "type": "image_generation",
            "prompt": "Premium studio product photo. Product: Ceramic mug",
            "seed": 1000,
            "width": 512,
            "height": 512,
            "quality_profile": "photo_studio_v1",
            "conditioning": {"controlnet_refs": ["refs/brand.png"]},
        }
    )
    asset_path = tmp_path / "images" / "variant-1000-watermarked.png"
    assert result["asset_path"] == str(asset_path)
    assert asset_path.is_file()


def test_local_client_generates_video_from_image(tmp_path) -> None:
    client = LocalGenerationClient(output_dir=tmp_path)
    image_result = client.run_workflow(
        {
            "type": "image_generation",
            "prompt": "Product shot",
            "seed": 2000,
            "width": 256,
            "height": 256,
            "quality_profile": "photo_studio_v1",
            "conditioning": {"controlnet_refs": []},
        }
    )
    video_result = client.run_workflow(
        {
            "type": "video_generation",
            "input_image_path": image_result["asset_path"],
            "duration_seconds": 2,
            "fps": 8,
            "seed": 2005,
        }
    )
    assert video_result["asset_path"].endswith((".mp4", ".gif"))
    assert Path(video_result["asset_path"]).is_file()


def test_is_comfyui_available_false_on_connection_error(monkeypatch) -> None:
    class _BrokenResponse:
        ok = False

    def _broken_get(*_args, **_kwargs):
        raise ConnectionError("offline")

    monkeypatch.setattr("ai_image_video_generator.pipelines.comfy_client.requests.get", _broken_get)
    assert is_comfyui_available("http://127.0.0.1:8188") is False
