"""Pytest fixtures: isolate in-memory store per test."""

import pytest

from app import main as app_main


@pytest.fixture(autouse=True)
def reset_store():
    """Reset the in-memory items store before each test so tests do not affect each other."""
    app_main._items.clear()
    app_main._next_id = 1
    yield
    app_main._items.clear()
    app_main._next_id = 1
