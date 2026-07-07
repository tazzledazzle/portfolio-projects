from pathlib import Path

from fastapi.testclient import TestClient

from ai_app.main import app
from ai_app.services.conversation import ConversationService
from ai_app.services.memory import MemoryService
from ai_app.services.tools import ToolService


def _test_client_with_stubbed_weather(tmp_path: Path) -> TestClient:
    memory = MemoryService(store_path=tmp_path / "memory.json")
    tools = ToolService()
    tools.register("weather", lambda city: f"Stub weather for {city}")
    app.dependency_overrides = {}
    import ai_app.main as main_module

    main_module.conversation_service = ConversationService(memory_service=memory, tool_service=tools)
    return TestClient(app)


def test_chat_direct_response_path(tmp_path) -> None:
    client = _test_client_with_stubbed_weather(tmp_path)
    response = client.post("/chat", json={"user_id": "u1", "message": "Hello assistant"})
    assert response.status_code == 200
    assert "You said: Hello assistant" in response.text


def test_chat_tool_response_path(tmp_path) -> None:
    client = _test_client_with_stubbed_weather(tmp_path)
    response = client.post("/chat", json={"user_id": "u1", "message": "weather in Paris"})
    assert response.status_code == 200
    assert "Stub weather for Paris" in response.text


def test_chat_memory_augmented_path(tmp_path) -> None:
    client = _test_client_with_stubbed_weather(tmp_path)
    first = client.post("/chat", json={"user_id": "u1", "message": "I enjoy hiking mountains"})
    assert first.status_code == 200

    second = client.post("/chat", json={"user_id": "u1", "message": "Any mountain tips?"})
    assert second.status_code == 200
    assert "I remember: I enjoy hiking mountains" in second.text
