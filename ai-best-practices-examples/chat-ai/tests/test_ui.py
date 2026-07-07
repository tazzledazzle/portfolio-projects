from fastapi.testclient import TestClient

from ai_app.main import app


def test_index_page_contains_chat_ui() -> None:
    client = TestClient(app)
    response = client.get("/")
    assert response.status_code == 200
    assert "AI Chat Assistant MVP" in response.text
    assert 'id="chat-form"' in response.text
    assert "fetch('/chat'" in response.text
