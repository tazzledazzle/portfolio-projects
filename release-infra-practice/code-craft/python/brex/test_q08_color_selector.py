"""Tests for Q8 — Color Multi-Select."""

import pytest

from q08_color_selector import ColorSelector


@pytest.fixture
def sel() -> ColorSelector:
    return ColorSelector()


def test_select_adds_color(sel: ColorSelector) -> None:
    sel.select("red")
    assert "red" in sel.selected


def test_select_invalid_color_raises(sel: ColorSelector) -> None:
    with pytest.raises(ValueError):
        sel.select("magenta")


def test_deselect_removes_color(sel: ColorSelector) -> None:
    sel.select("blue")
    sel.deselect("blue")
    assert "blue" not in sel.selected


def test_deselect_noop_when_not_selected(sel: ColorSelector) -> None:
    sel.deselect("green")
    assert sel.selected == []


def test_toggle_selects_when_unselected(sel: ColorSelector) -> None:
    sel.toggle("yellow")
    assert "yellow" in sel.selected


def test_toggle_deselects_when_selected(sel: ColorSelector) -> None:
    sel.select("purple")
    sel.toggle("purple")
    assert "purple" not in sel.selected


def test_selected_returns_sorted(sel: ColorSelector) -> None:
    sel.select("red")
    sel.select("blue")
    assert sel.selected == ["blue", "red"]


def test_multi_select(sel: ColorSelector) -> None:
    sel.select("red")
    sel.select("blue")
    sel.select("green")
    assert len(sel.selected) == 3


def test_duplicate_select_idempotent(sel: ColorSelector) -> None:
    sel.select("orange")
    sel.select("orange")
    assert sel.selected.count("orange") == 1


def test_filter_palette_matches_substring(sel: ColorSelector) -> None:
    results = sel.filter_palette("re")
    assert "red" in results
    assert "green" in results


def test_filter_palette_case_insensitive(sel: ColorSelector) -> None:
    results = sel.filter_palette("RED")
    assert "red" in results


def test_filter_palette_no_match_returns_empty(sel: ColorSelector) -> None:
    results = sel.filter_palette("zzz")
    assert results == []


def test_selected_details_returns_metadata(sel: ColorSelector) -> None:
    sel.select("red")
    details = sel.selected_details()
    assert details[0]["hex"] == "#FF0000"


def test_selected_details_sorted_by_name(sel: ColorSelector) -> None:
    sel.select("red")
    sel.select("blue")
    details = sel.selected_details()
    assert details[0]["hex"] == "#0000FF"


def test_selected_details_empty_when_none_selected(sel: ColorSelector) -> None:
    assert sel.selected_details() == []
