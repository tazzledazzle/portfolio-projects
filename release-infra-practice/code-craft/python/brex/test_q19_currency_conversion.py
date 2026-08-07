"""Tests for Q19 — Multi-currency Conversion with Correct Rounding."""

from decimal import Decimal

import pytest

from q19_currency_conversion import BatchItem, ConversionResult, batch_convert, convert

RATES: dict[str, Decimal] = {
    "USD/EUR": Decimal("0.92"),
    "USD/JPY": Decimal("149.50"),
    "EUR/GBP": Decimal("0.86"),
    "USD/KWD": Decimal("0.307"),
}


def test_convert_usd_to_eur() -> None:
    result = convert(Decimal("100"), "USD", "EUR", RATES)
    assert result.converted_amount == Decimal("92.00")


def test_convert_rounds_to_currency_precision() -> None:
    result = convert(Decimal("1"), "USD", "EUR", RATES)
    assert result.converted_amount == Decimal("0.92")


def test_convert_jpy_no_decimals() -> None:
    result = convert(Decimal("100"), "USD", "JPY", RATES)
    assert result.converted_amount == Decimal("14950")


def test_convert_uses_inverse_rate() -> None:
    result = convert(Decimal("92"), "EUR", "USD", RATES)
    assert result.from_currency == "EUR"
    assert result.to_currency == "USD"


def test_convert_unknown_rate_raises() -> None:
    with pytest.raises(KeyError):
        convert(Decimal("100"), "JPY", "KWD", RATES)


def test_convert_unknown_currency_raises() -> None:
    with pytest.raises(ValueError):
        convert(Decimal("100"), "USD", "ZZZ", RATES)


def test_convert_result_contains_original_amount() -> None:
    result = convert(Decimal("50"), "USD", "EUR", RATES)
    assert result.original_amount == Decimal("50")


def test_convert_kwd_three_decimal_precision() -> None:
    result = convert(Decimal("10"), "USD", "KWD", RATES)
    decimal_places = result.converted_amount.as_tuple().exponent
    assert decimal_places >= -3


def test_convert_same_currency_identity() -> None:
    rates_with_identity: dict[str, Decimal] = {**RATES, "USD/USD": Decimal("1")}
    result = convert(Decimal("100"), "USD", "USD", rates_with_identity)
    assert result.converted_amount == Decimal("100.00")


def test_batch_convert_all_success() -> None:
    items = [
        BatchItem(Decimal("100"), "USD", "EUR"),
        BatchItem(Decimal("50"), "USD", "JPY"),
    ]
    results = batch_convert(items, RATES)
    assert all(r.error is None for r in results)
    assert len(results) == 2


def test_batch_convert_partial_error() -> None:
    items = [
        BatchItem(Decimal("100"), "USD", "EUR"),
        BatchItem(Decimal("100"), "JPY", "KWD"),
    ]
    results = batch_convert(items, RATES)
    successes = [r for r in results if r.error is None]
    errors = [r for r in results if r.error is not None]
    assert len(successes) == 1
    assert len(errors) == 1


def test_batch_convert_error_does_not_abort() -> None:
    items = [
        BatchItem(Decimal("100"), "JPY", "KWD"),
        BatchItem(Decimal("100"), "USD", "EUR"),
    ]
    results = batch_convert(items, RATES)
    assert len(results) == 2


def test_batch_convert_result_order_matches_input() -> None:
    items = [
        BatchItem(Decimal("10"), "USD", "EUR"),
        BatchItem(Decimal("20"), "USD", "EUR"),
    ]
    results = batch_convert(items, RATES)
    assert results[0].item.amount == Decimal("10")
    assert results[1].item.amount == Decimal("20")


def test_batch_convert_error_sets_error_field() -> None:
    items = [BatchItem(Decimal("100"), "JPY", "KWD")]
    results = batch_convert(items, RATES)
    assert results[0].error is not None
    assert results[0].result is None
