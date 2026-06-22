from src.classify.compatibility_rules import classify_change


def test_removed_path_is_breaking() -> None:
    assert classify_change({"kind": "path_removed"}) == "breaking"
