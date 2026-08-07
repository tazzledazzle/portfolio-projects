"""Tests for Q14 — P2P Payment App."""

from decimal import Decimal

import pytest

from q14_p2p_payment import InsufficientFundsError, P2PService, PaymentStatus


@pytest.fixture
def svc() -> P2PService:
    s = P2PService()
    s.register("alice", "Alice", Decimal("500"))
    s.register("bob", "Bob", Decimal("100"))
    return s


def test_register_user(svc: P2PService) -> None:
    user = svc.register("carol", "Carol", Decimal("200"))
    assert user.user_id == "carol"


def test_register_duplicate_raises(svc: P2PService) -> None:
    with pytest.raises(ValueError):
        svc.register("alice", "Alice 2")


def test_send_moves_funds(svc: P2PService) -> None:
    svc.send("p1", "alice", "bob", Decimal("100"))
    assert svc._users["alice"].balance == Decimal("400")
    assert svc._users["bob"].balance == Decimal("200")


def test_send_returns_completed_payment(svc: P2PService) -> None:
    pmt = svc.send("p2", "alice", "bob", Decimal("50"))
    assert pmt.status == PaymentStatus.COMPLETED


def test_send_insufficient_funds_raises(svc: P2PService) -> None:
    with pytest.raises(InsufficientFundsError):
        svc.send("p3", "bob", "alice", Decimal("999"))


def test_send_zero_amount_raises(svc: P2PService) -> None:
    with pytest.raises(ValueError):
        svc.send("p4", "alice", "bob", Decimal("0"))


def test_send_unknown_user_raises(svc: P2PService) -> None:
    with pytest.raises(KeyError):
        svc.send("p5", "ghost", "bob", Decimal("10"))


def test_request_creates_pending_payment(svc: P2PService) -> None:
    pmt = svc.request("p6", "bob", "alice", Decimal("50"))
    assert pmt.status == PaymentStatus.PENDING


def test_request_does_not_move_funds(svc: P2PService) -> None:
    svc.request("p7", "bob", "alice", Decimal("50"))
    assert svc._users["alice"].balance == Decimal("500")


def test_accept_request_moves_funds(svc: P2PService) -> None:
    svc.request("p8", "bob", "alice", Decimal("50"))
    svc.accept_request("p8")
    assert svc._users["alice"].balance == Decimal("450")
    assert svc._users["bob"].balance == Decimal("150")


def test_decline_request_no_funds_moved(svc: P2PService) -> None:
    svc.request("p9", "bob", "alice", Decimal("50"))
    pmt = svc.decline_request("p9")
    assert pmt.status == PaymentStatus.DECLINED
    assert svc._users["alice"].balance == Decimal("500")


def test_history_includes_sent_and_received(svc: P2PService) -> None:
    svc.send("p10", "alice", "bob", Decimal("10"))
    h = svc.history("alice")
    assert any(p.payment_id == "p10" for p in h)


def test_history_ordered_oldest_first(svc: P2PService) -> None:
    svc.send("p11", "alice", "bob", Decimal("10"))
    svc.send("p12", "alice", "bob", Decimal("20"))
    h = svc.history("alice")
    assert h[0].payment_id == "p11"
