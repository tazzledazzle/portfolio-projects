from ai_image_video_generator.pipelines.comfy_client import ComfyClient
from ai_image_video_generator.pipelines.comfy_client import ComfyClientError


class FakeTransport:
    def __init__(self) -> None:
        self.last_path = ""
        self.last_payload = {}

    def post_json(self, path: str, payload: dict) -> dict:
        self.last_path = path
        self.last_payload = payload
        return {
            "asset_path": "outputs/images/test.png",
            "seed": payload.get("seed", 42),
            "model_name": "sdxl-base",
            "workflow_name": "image_generation",
            "created_at": "2026-03-31T12:00:00+00:00",
        }


def test_comfy_client_submits_workflow() -> None:
    transport = FakeTransport()
    client = ComfyClient(base_url="http://127.0.0.1:8188", transport=transport)
    result = client.run_workflow({"seed": 42, "prompt": "test prompt"})
    assert result["asset_path"].endswith("test.png")
    assert transport.last_path == "/api/generate"


def test_comfy_client_raises_on_invalid_response() -> None:
    class BadTransport:
        def post_json(self, path: str, payload: dict) -> dict:
            return {}

    client = ComfyClient(base_url="http://127.0.0.1:8188", transport=BadTransport())
    try:
        client.run_workflow({"seed": 1})
    except ComfyClientError as exc:
        assert "asset_path" in str(exc)
    else:
        raise AssertionError("Expected ComfyClientError for malformed response")
