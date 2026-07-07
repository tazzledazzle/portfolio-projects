from src.classify.compatibility_rules import classify_change


def should_fail(changes: list[dict]) -> bool:
    return any(classify_change(change) == "breaking" for change in changes)
