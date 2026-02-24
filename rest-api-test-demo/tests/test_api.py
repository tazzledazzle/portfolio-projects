"""Integration tests for the API using FastAPI TestClient."""

import pytest
from fastapi.testclient import TestClient

from app.main import app

client = TestClient(app)


def test_health() -> None:
    r = client.get("/health")
    assert r.status_code == 200
    assert r.json() == {"status": "ok"}


def test_list_items_empty() -> None:
    r = client.get("/items")
    assert r.status_code == 200
    assert r.json() == []


def test_create_and_get_item() -> None:
    r = client.post("/items", json={"name": "Test", "description": "A test item"})
    assert r.status_code == 201
    data = r.json()
    assert data["name"] == "Test"
    assert data["description"] == "A test item"
    assert "id" in data
    item_id = data["id"]
    r2 = client.get(f"/items/{item_id}")
    assert r2.status_code == 200
    assert r2.json()["name"] == "Test"


def test_get_item_not_found() -> None:
    r = client.get("/items/99999")
    assert r.status_code == 404
    assert "not found" in r.json()["detail"].lower()


def test_delete_item() -> None:
    r = client.post("/items", json={"name": "To delete"})
    assert r.status_code == 201
    item_id = r.json()["id"]
    r2 = client.delete(f"/items/{item_id}")
    assert r2.status_code == 204
    r3 = client.get(f"/items/{item_id}")
    assert r3.status_code == 404


def test_delete_item_not_found() -> None:
    r = client.delete("/items/99999")
    assert r.status_code == 404
