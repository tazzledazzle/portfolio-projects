"""Tests for Workflow API: health and models. Mocks get_redis/get_db so no real services needed."""

from unittest.mock import AsyncMock, MagicMock, patch
import pytest
from fastapi.testclient import TestClient
from pydantic import ValidationError

from main import JobCreate, JobResponse, app


@pytest.fixture
def fake_redis():
    """In-memory Redis-like for tests."""
    class Fake:
        def __init__(self):
            self._list = []
            self._store = {}
        async def lpush(self, key, value):
            self._list.insert(0, value)
        async def set(self, key, value):
            self._store[key] = value
        async def get(self, key):
            return self._store.get(key)
    return Fake()


@pytest.fixture
def fake_pool():
    """Fake asyncpg pool with in-memory jobs dict."""
    jobs = {}
    class FakeConn:
        def __init__(self):
            self.jobs = jobs
        async def execute(self, query, *args):
            if "CREATE TABLE" in query:
                return
            if "INSERT INTO jobs" in query:
                self.jobs[args[0]] = {"id": args[0], "status": args[1], "result": None}
        async def fetchrow(self, query, job_id):
            return self.jobs.get(job_id)
        async def __aenter__(self):
            return self
        async def __aexit__(self, *args):
            pass
    class FakePool:
        async def acquire(self):
            return FakeConn()
    return FakePool()


def test_health(fake_redis, fake_pool):
    with patch("main.get_redis", new_callable=AsyncMock, return_value=fake_redis):
        with patch("main.get_db", new_callable=AsyncMock, return_value=fake_pool):
            client = TestClient(app)
            r = client.get("/health")
    assert r.status_code == 200
    assert r.json() == {"status": "ok"}


def test_job_create_model() -> None:
    JobCreate(payload="hello")
    with pytest.raises(ValidationError):
        JobCreate()


def test_job_response_model() -> None:
    JobResponse(id="x", status="queued", result=None)
    JobResponse(id="y", status="completed", result="done")
