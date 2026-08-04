"""Tests for Q3 — Gem Card Purchases with Color Discounts."""

from decimal import Decimal

import pytest

from q03_gem_card_purchases import GemCard, GemColor, apply_discount


def test_red_discount_five_percent() -> None:
    result = apply_discount(GemColor.RED, Decimal("100"))
    assert result == Decimal("95")


def test_blue_discount_ten_percent() -> None:
    result = apply_discount(GemColor.BLUE, Decimal("100"))
    assert result == Decimal("90")


def test_green_discount_fifteen_percent() -> None:
    result = apply_discount(GemColor.GREEN, Decimal("100"))
    assert result == Decimal("85")


def test_yellow_discount_twenty_percent() -> None:
    result = apply_discount(GemColor.YELLOW, Decimal("100"))
    assert result == Decimal("80")


def test_purple_discount_twenty_five_percent() -> None:
    result = apply_discount(GemColor.PURPLE, Decimal("100"))
    assert result == Decimal("75")


def test_gem_card_add_gem() -> None:
    card = GemCard("g1")
    card.add_gem(GemColor.BLUE, 2)
    assert card.total_gems == 2


def test_best_discount_picks_highest_discount() -> None:
    card = GemCard("g2")
    card.add_gem(GemColor.RED)
    card.add_gem(GemColor.PURPLE)
    price = card.best_discount_price(Decimal("100"))
    assert price == Decimal("75")


def test_best_discount_no_gems_raises() -> None:
    card = GemCard("g3")
    with pytest.raises(ValueError):
        card.best_discount_price(Decimal("100"))


def test_purchase_consumes_gem() -> None:
    card = GemCard("g4")
    card.add_gem(GemColor.GREEN)
    card.purchase_with_best_gem(Decimal("100"))
    assert card.total_gems == 0


def test_purchase_returns_correct_price_and_color() -> None:
    card = GemCard("g5")
    card.add_gem(GemColor.YELLOW)
    price, color = card.purchase_with_best_gem(Decimal("200"))
    assert price == Decimal("160")
    assert color == GemColor.YELLOW


def test_purchase_no_gems_raises() -> None:
    card = GemCard("g6")
    with pytest.raises(ValueError):
        card.purchase_with_best_gem(Decimal("50"))


def test_multiple_gems_same_color() -> None:
    card = GemCard("g7")
    card.add_gem(GemColor.BLUE, 3)
    card.purchase_with_best_gem(Decimal("100"))
    assert card.total_gems == 2


def test_discount_on_fractional_amount() -> None:
    result = apply_discount(GemColor.BLUE, Decimal("33.33"))
    assert result == Decimal("33.33") * Decimal("0.90")


def test_add_gem_accumulates() -> None:
    card = GemCard("g8")
    card.add_gem(GemColor.RED, 1)
    card.add_gem(GemColor.RED, 2)
    assert card.gems.get(GemColor.RED, 0) == 3


def test_best_price_single_gem_color() -> None:
    card = GemCard("g9")
    card.add_gem(GemColor.GREEN)
    assert card.best_discount_price(Decimal("100")) == Decimal("85")
