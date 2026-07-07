from fastapi.testclient import TestClient

from ws_chat.main import app


def test_health_route_not_required_login_page_exists() -> None:
    client = TestClient(app)
    response = client.get("/login")
    assert response.status_code == 200
