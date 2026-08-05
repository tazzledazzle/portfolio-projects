"""
Q3 — Gem Card Purchases with Color Discounts

Five gem colors each carry a distinct discount rate applied to a purchase amount.
Base: given a color and purchase amount, return the discounted price.
Ext 1: a card holds gems of multiple colors; the cheapest applicable discount wins.
Ext 2: gems are consumed on purchase (each gem has one use); track remaining gems.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from decimal import Decimal
from enum import Enum


class GemColor(Enum):
    RED = "red"
    BLUE = "blue"
    GREEN = "green"
    YELLOW = "yellow"
    PURPLE = "purple"


# Discount rates per color (as a fraction, e.g. 0.10 == 10% off)
GEM_DISCOUNT: dict[GemColor, Decimal] = {
    GemColor.RED: Decimal("0.05"),
    GemColor.BLUE: Decimal("0.10"),
    GemColor.GREEN: Decimal("0.15"),
    GemColor.YELLOW: Decimal("0.20"),
    GemColor.PURPLE: Decimal("0.25"),
}


def apply_discount(color: GemColor, amount: Decimal) -> Decimal:
    """Return price after applying the gem color's discount to amount."""
    return amount - (amount * GEM_DISCOUNT[color])


@dataclass
class GemCard:
    card_id: str
    gems: dict[GemColor, int] = field(default_factory=dict)

    def add_gem(self, color: GemColor, count: int = 1) -> None:
        """Add gems of the given color to the card."""
        if color not in self.gems:
            self.gems[color] = count
        else:
            self.gems[color] += count

    def best_discount_price(self, amount: Decimal) -> Decimal:
        """
        Return the lowest price achievable using any available gem.
        Raises ValueError if the card has no gems.
        """
        discounts = {}
        for color, quantity in self.gems.items():
            # print(f'{color}: {quantity} -> {(amount * quantity) - apply_discount(color, amount) * quantity} ')
            discounts[color] = apply_discount(color, amount)
            ## need inventory of gems
        best_color = min(discounts, key=discounts.get)

        return apply_discount(best_color, amount)

    def purchase_with_best_gem(self, amount: Decimal) -> tuple[Decimal, GemColor]:
        """
        Apply the best discount, consume one gem of that color, and return
        (discounted_price, gem_color_used).
        Raises ValueError if no gems remain.
        """
        if not self.total_gems:
            raise ValueError("No gems available.")
        best_color = min(
            (color for color, quantity in self.gems.items() if quantity > 0 ),
            key=lambda color: apply_discount(color, amount),
        )
        self.gems[best_color] -= 1
        return apply_discount(best_color, amount), best_color


    @property
    def total_gems(self) -> int:
        return sum([x for x in self.gems.values()])


