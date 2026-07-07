from friction_scorer.core.composite import compute_composite_score


def test_compute_composite_score() -> None:
    result = compute_composite_score(
        {
            "setup_time": 2.0,
            "build_time": 2.0,
            "test_time": 2.0,
            "review_latency": 2.0,
            "deploy_time": 2.0,
        }
    )
    assert result == 2.0
