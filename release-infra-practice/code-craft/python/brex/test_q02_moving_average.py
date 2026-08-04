"""Tests for Q2 — Moving Average of Transaction Volumes."""

from decimal import Decimal

import pytest

from q02_moving_average import MovingAverage


@pytest.fixture
def ma3() -> MovingAverage:
    return MovingAverage(window_size=3)


def test_single_element_average(ma3: MovingAverage) -> None:
    result = ma3.add(Decimal("100"))
    assert result == Decimal("100")


def test_two_elements_average(ma3: MovingAverage) -> None:
    ma3.add(Decimal("100"))
    result = ma3.add(Decimal("200"))
    assert result == Decimal("150")


def test_full_window_average(ma3: MovingAverage) -> None:
    ma3.add(Decimal("100"))
    ma3.add(Decimal("200"))
    result = ma3.add(Decimal("300"))
    assert result == Decimal("200")


def test_window_slides_old_value_removed(ma3: MovingAverage) -> None:
    ma3.add(Decimal("100"))
    ma3.add(Decimal("200"))
    ma3.add(Decimal("300"))
    result = ma3.add(Decimal("400"))
    assert result == Decimal("300")


def test_current_average_property(ma3: MovingAverage) -> None:
    ma3.add(Decimal("50"))
    ma3.add(Decimal("150"))
    assert ma3.current_average == Decimal("100")


def test_global_total_tracks_all_inserts(ma3: MovingAverage) -> None:
    ma3.add(Decimal("100"))
    ma3.add(Decimal("200"))
    ma3.add(Decimal("300"))
    ma3.add(Decimal("400"))
    assert ma3.global_total == Decimal("1000")


def test_reset_clears_window(ma3: MovingAverage) -> None:
    ma3.add(Decimal("100"))
    ma3.add(Decimal("200"))
    ma3.reset()
    ma3.add(Decimal("50"))
    assert ma3.current_average == Decimal("50")


def test_reset_preserves_global_total(ma3: MovingAverage) -> None:
    ma3.add(Decimal("300"))
    ma3.reset()
    assert ma3.global_total == Decimal("300")


def test_invalid_window_size_raises() -> None:
    with pytest.raises(ValueError):
        MovingAverage(window_size=0)


def test_window_size_one(ma1: MovingAverage = MovingAverage(1)) -> None:
    ma1.add(Decimal("10"))
    result = ma1.add(Decimal("20"))
    assert result == Decimal("20")


def test_average_precision(ma3: MovingAverage) -> None:
    ma3.add(Decimal("1"))
    ma3.add(Decimal("2"))
    result = ma3.add(Decimal("3"))
    assert result == Decimal("2")


def test_zero_value_included(ma3: MovingAverage) -> None:
    ma3.add(Decimal("0"))
    ma3.add(Decimal("0"))
    result = ma3.add(Decimal("0"))
    assert result == Decimal("0")


def test_large_values(ma3: MovingAverage) -> None:
    ma3.add(Decimal("1000000"))
    ma3.add(Decimal("2000000"))
    result = ma3.add(Decimal("3000000"))
    assert result == Decimal("2000000")


def test_add_returns_decimal(ma3: MovingAverage) -> None:
    result = ma3.add(Decimal("5"))
    assert isinstance(result, Decimal)


def test_window_never_exceeds_k(ma3: MovingAverage) -> None:
    for i in range(10):
        ma3.add(Decimal(str(i)))
    assert len(ma3._window) <= 3
