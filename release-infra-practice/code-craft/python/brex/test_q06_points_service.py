"""Tests for Q6 — Points-from-Transactions Service."""

from decimal import Decimal

import pytest

from q06_points_service import POINTS_PER_USD, PointsAccount


@pytest.fixture
def acct() -> PointsAccount:
    return PointsAccount(account_id="p1")


def test_initial_balance_zero(acct: PointsAccount) -> None:
    assert acct.balance == Decimal("0")


def test_award_one_dollar(acct: PointsAccount) -> None:
    awarded = acct.award(Decimal("1"), "k1")
    assert awarded == Decimal("1.5")


def test_award_increases_balance(acct: PointsAccount) -> None:
    acct.award(Decimal("10"), "k2")
    assert acct.balance == Decimal("15")


def test_award_idempotent_duplicate_key(acct: PointsAccount) -> None:
    acct.award(Decimal("10"), "k3")
    awarded_again = acct.award(Decimal("10"), "k3")
    assert awarded_again == Decimal("0")
    assert acct.balance == Decimal("15")


def test_redeem_reduces_balance(acct: PointsAccount) -> None:
    acct.award(Decimal("20"), "k4")
    acct.redeem(Decimal("10"), "r1")
    assert acct.balance == Decimal("20")


def test_redeem_returns_true_on_success(acct: PointsAccount) -> None:
    acct.award(Decimal("10"), "k5")
    assert acct.redeem(Decimal("5"), "r2") is True


def test_redeem_returns_false_insufficient_balance(acct: PointsAccount) -> None:
    assert acct.redeem(Decimal("100"), "r3") is False


def test_redeem_no_double_spend(acct: PointsAccount) -> None:
    acct.award(Decimal("10"), "k6")
    acct.redeem(Decimal("15"), "r4")
    assert acct.balance == Decimal("15")


def test_redeem_idempotent_duplicate_key(acct: PointsAccount) -> None:
    acct.award(Decimal("20"), "k7")
    acct.redeem(Decimal("5"), "r5")
    result = acct.redeem(Decimal("5"), "r5")
    assert result is False
    assert acct.balance == Decimal("25")


def test_ledger_sum_matches_balance(acct: PointsAccount) -> None:
    acct.award(Decimal("10"), "k8")
    acct.award(Decimal("5"), "k9")
    acct.redeem(Decimal("3"), "r6")
    assert acct.ledger_sum() == acct.balance


def test_points_per_usd_constant() -> None:
    assert POINTS_PER_USD == Decimal("1.5")


def test_award_fractional_usd(acct: PointsAccount) -> None:
    acct.award(Decimal("0.50"), "k10")
    assert acct.balance == Decimal("0.75")


def test_multiple_awards_accumulate(acct: PointsAccount) -> None:
    acct.award(Decimal("1"), "k11")
    acct.award(Decimal("1"), "k12")
    assert acct.balance == Decimal("3.0")


def test_full_redemption_leaves_zero(acct: PointsAccount) -> None:
    acct.award(Decimal("10"), "k13")
    acct.redeem(Decimal("15"), "r7")
    assert acct.balance == Decimal("0")


def test_ledger_empty_initially(acct: PointsAccount) -> None:
    assert acct.ledger_sum() == Decimal("0")
