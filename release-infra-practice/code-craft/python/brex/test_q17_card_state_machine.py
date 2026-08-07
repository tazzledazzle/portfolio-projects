"""Tests for Q17 — Card Freeze/Unfreeze State Machine."""

import pytest

from q17_card_state_machine import CardState, ManagedCard


@pytest.fixture
def card() -> ManagedCard:
    return ManagedCard(card_id="card-1")


def test_initial_state_active(card: ManagedCard) -> None:
    assert card.state == CardState.ACTIVE


def test_freeze_transitions_to_frozen(card: ManagedCard) -> None:
    card.freeze("admin")
    assert card.state == CardState.FROZEN


def test_unfreeze_transitions_to_active(card: ManagedCard) -> None:
    card.freeze("admin")
    card.unfreeze("admin")
    assert card.state == CardState.ACTIVE


def test_freeze_twice_raises(card: ManagedCard) -> None:
    card.freeze("admin")
    with pytest.raises(ValueError):
        card.freeze("admin")


def test_unfreeze_active_raises(card: ManagedCard) -> None:
    with pytest.raises(ValueError):
        card.unfreeze("admin")


def test_close_from_active(card: ManagedCard) -> None:
    card.close("admin")
    assert card.state == CardState.CLOSED


def test_close_from_frozen(card: ManagedCard) -> None:
    card.freeze("admin")
    card.close("admin")
    assert card.state == CardState.CLOSED


def test_close_twice_raises(card: ManagedCard) -> None:
    card.close("admin")
    with pytest.raises(ValueError):
        card.close("admin")


def test_freeze_closed_raises(card: ManagedCard) -> None:
    card.close("admin")
    with pytest.raises(ValueError):
        card.freeze("admin")


def test_unfreeze_closed_raises(card: ManagedCard) -> None:
    card.freeze("admin")
    card.close("admin")
    with pytest.raises(ValueError):
        card.unfreeze("admin")


def test_audit_trail_populated_on_freeze(card: ManagedCard) -> None:
    card.freeze("admin")
    trail = card.audit_trail
    assert len(trail) == 1
    assert trail[0].from_state == CardState.ACTIVE
    assert trail[0].to_state == CardState.FROZEN


def test_audit_trail_records_actor(card: ManagedCard) -> None:
    card.freeze("user-99")
    assert card.audit_trail[0].actor == "user-99"


def test_audit_trail_is_copy(card: ManagedCard) -> None:
    card.freeze("admin")
    trail = card.audit_trail
    trail.clear()
    assert len(card.audit_trail) == 1


def test_audit_trail_multiple_transitions(card: ManagedCard) -> None:
    card.freeze("admin")
    card.unfreeze("admin")
    card.close("admin")
    assert len(card.audit_trail) == 3


def test_audit_trail_close_entry(card: ManagedCard) -> None:
    card.close("admin")
    assert card.audit_trail[0].to_state == CardState.CLOSED
