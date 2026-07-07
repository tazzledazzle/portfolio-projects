from app.services.scoring import rolling_nps


def test_rolling_nps_mixed_responses() -> None:
    result = rolling_nps([10, 9, 8, 7, 6, 4])
    assert result == 0.0
