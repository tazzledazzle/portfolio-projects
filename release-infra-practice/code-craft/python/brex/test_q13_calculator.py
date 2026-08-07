"""Tests for Q13 — Calculator."""

import pytest

from q13_calculator import calculate


def test_addition() -> None:
    assert calculate("1 + 2") == 3


def test_subtraction() -> None:
    assert calculate("10 - 4") == 6


def test_multiplication() -> None:
    assert calculate("3 * 4") == 12


def test_integer_division() -> None:
    assert calculate("10 / 3") == 3


def test_operator_precedence_mul_before_add() -> None:
    assert calculate("2 + 3 * 4") == 14


def test_operator_precedence_div_before_sub() -> None:
    assert calculate("10 - 6 / 2") == 7


def test_parentheses_override_precedence() -> None:
    assert calculate("(2 + 3) * 4") == 20


def test_nested_parentheses() -> None:
    assert calculate("((2 + 3) * (1 + 1))") == 10


def test_unary_negation() -> None:
    assert calculate("-3 + 5") == 2


def test_unary_negation_in_expression() -> None:
    assert calculate("5 * -2") == -10


def test_whitespace_ignored() -> None:
    assert calculate("  3  +  4  ") == 7


def test_division_truncates_toward_zero_negative() -> None:
    assert calculate("-7 / 2") == -3


def test_division_by_zero_raises() -> None:
    with pytest.raises(ZeroDivisionError):
        calculate("5 / 0")


def test_complex_expression() -> None:
    assert calculate("3 + 2 * 2") == 7


def test_single_number() -> None:
    assert calculate("42") == 42
