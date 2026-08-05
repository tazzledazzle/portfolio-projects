"""Tests for Q4 — Credit-limit Checker."""

from decimal import Decimal

import pytest

from q04_credit_limit_checker import Account, CreditService


@pytest.fixture
def account() -> Account:
    return Account(account_id="a1", credit_limit=Decimal("500"))


@pytest.fixture
def service(account: Account) -> CreditService:
    svc = CreditService()
    svc.add_account(account)
    return svc


def test_can_charge_within_limit(account: Account) -> None:
    assert account.can_charge(Decimal("300")) is True


def test_can_charge_exact_limit(account: Account) -> None:
    assert account.can_charge(Decimal("500")) is True


def test_can_charge_over_limit(account: Account) -> None:
    assert account.can_charge(Decimal("501")) is False


def test_charge_reduces_balance(account: Account) -> None:
    account.charge(Decimal("200"))
    assert account.balance == Decimal("200")


def test_charge_returns_false_when_declined(account: Account) -> None:
    assert account.charge(Decimal("9999")) is False


def test_available_credit(account: Account) -> None:
    account.charge(Decimal("100"))
    assert account.available_credit == Decimal("400")


def test_override_limit_allows_larger_charge(account: Account) -> None:
    account.set_override_limit(Decimal("800"))
    assert account.can_charge(Decimal("750")) is True


def test_override_limit_consumed_after_use(account: Account) -> None:
    account.set_override_limit(Decimal("800"))
    account.charge(Decimal("600"))
    assert account._override_limit is None


def test_override_not_consumed_if_charge_within_normal_limit(account: Account) -> None:
    account.set_override_limit(Decimal("800"))
    account.charge(Decimal("100"))
    assert account._override_limit == Decimal("800")


def test_service_get_account(service: CreditService, account: Account) -> None:
    retrieved = service.get_account("a1")
    assert retrieved is account


def test_service_get_missing_account_raises(service: CreditService) -> None:
    with pytest.raises(KeyError):
        service.get_account("missing")


def test_service_can_charge(service: CreditService) -> None:
    assert service.can_charge("a1", Decimal("100")) is True


def test_service_charge_updates_balance(service: CreditService) -> None:
    service.charge("a1", Decimal("250"))
    assert service.get_account("a1").balance == Decimal("250")


def test_service_charge_declined_over_limit(service: CreditService) -> None:
    assert service.charge("a1", Decimal("9999")) is False


def test_balance_unchanged_on_declined_charge(account: Account) -> None:
    account.charge(Decimal("9999"))
    assert account.balance == Decimal("0")
