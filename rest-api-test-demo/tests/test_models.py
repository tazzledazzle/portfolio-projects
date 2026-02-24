"""Unit tests for request/response validation and models."""

import pytest
from pydantic import ValidationError

from app.main import ItemCreate, Item


def test_item_create_valid() -> None:
    obj = ItemCreate(name="x", description="y")
    assert obj.name == "x"
    assert obj.description == "y"


def test_item_create_description_optional() -> None:
    obj = ItemCreate(name="x")
    assert obj.description is None


def test_item_create_name_required() -> None:
    with pytest.raises(ValidationError):
        ItemCreate(description="only")  # type: ignore


def test_item_valid() -> None:
    obj = Item(id=1, name="a", description="b")
    assert obj.id == 1
    assert obj.name == "a"
