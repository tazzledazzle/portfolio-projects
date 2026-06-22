from fastapi.testclient import TestClient

from ai_app.main import app


def test_health_endpoint_exists() -> None:
    client = TestClient(app)
    response = client.get("/health")
    assert response.status_code == 200
    assert response.json() == {"status": "ok"}


def test_chat_requires_payload() -> None:
    client = TestClient(app)
    response = client.post("/chat", json={})
    assert response.status_code == 422
