def classify_change(change: dict) -> str:
    if change["kind"] in {"path_removed", "required_field_added"}:
        return "breaking"
    return "non_breaking"
