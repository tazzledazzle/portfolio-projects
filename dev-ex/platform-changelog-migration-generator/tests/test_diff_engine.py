from changelog_generator.diff_engine import diff_api


def test_diff_api_returns_expected_categories() -> None:
    result = diff_api("old.json", "new.json")
    assert "breaking_changes" in result
    assert "deprecations" in result
    assert "new_features" in result
