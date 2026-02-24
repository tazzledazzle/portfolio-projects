"""Tests for OTel Demo API endpoints."""

import pytest
from fastapi.testclient import TestClient

# Import app after optional env to avoid OTLP connection issues in tests
from main import app

client = TestClient(app)


def test_health() -> None:
    r = client.get("/health")
    assert r.status_code == 200
    assert r.json() == {"status": "ok"}


def test_health_otel() -> None:
    r = client.get("/api/health/otel")
    assert r.status_code == 200
    data = r.json()
    assert data["status"] == "healthy"
    assert data["sdk"]["initialized"] is True
    assert "serviceName" in data["sdk"]
    assert len(data["exporters"]) >= 1
    assert data["testSpan"]["created"] is True


def test_example() -> None:
    r = client.get("/api/example")
    assert r.status_code == 200
    assert r.json() == {"message": "Hello from instrumented API"}
