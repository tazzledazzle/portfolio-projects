"""
Q19 — Multi-currency Conversion with Correct Rounding

No float arithmetic. All conversions use Decimal with explicit rounding rules.
Base: convert an amount in one currency to another using a rate table.
Ext 1: round to the target currency's precision (e.g. JPY = 0 decimals, USD = 2).
Ext 2: batch-convert a list of amounts, returning per-item results and any errors.
"""

from __future__ import annotations

from dataclasses import dataclass
from decimal import Decimal, ROUND_HALF_UP

# ISO 4217 decimal places per currency
CURRENCY_PRECISION: dict[str, int] = {
    "USD": 2,
    "EUR": 2,
    "GBP": 2,
    "JPY": 0,
    "CHF": 2,
    "KWD": 3,
}


@dataclass
class ConversionResult:
    from_currency: str
    to_currency: str
    original_amount: Decimal
    converted_amount: Decimal


@dataclass
class BatchItem:
    amount: Decimal
    from_currency: str
    to_currency: str


@dataclass
class BatchResult:
    item: BatchItem
    result: ConversionResult | None
    error: str | None = None


def convert(
    amount: Decimal,
    from_currency: str,
    to_currency: str,
    rates: dict[str, Decimal],
) -> ConversionResult:
    """
    Convert amount using rates[f"{from_currency}/{to_currency}"] or the inverse.

    Rounds to target currency precision (ROUND_HALF_UP).
    Raises:
        KeyError: if neither direct nor inverse rate exists.
        ValueError: if from_currency or to_currency unknown in CURRENCY_PRECISION.
    """
    raise NotImplementedError


def batch_convert(
    items: list[BatchItem],
    rates: dict[str, Decimal],
) -> list[BatchResult]:
    """Convert each item; errors produce a BatchResult with error set, not an exception."""
    raise NotImplementedError
