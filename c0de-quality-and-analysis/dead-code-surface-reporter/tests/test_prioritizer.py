from src.scoring.prioritizer import priority_score


def test_priority_score_bounds() -> None:
    score = priority_score(0.9, 180)
    assert 0.0 <= score <= 1.0
