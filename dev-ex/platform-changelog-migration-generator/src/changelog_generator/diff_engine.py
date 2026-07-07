def diff_api(old_spec_path: str, new_spec_path: str) -> dict[str, list[str]]:
    return {
        "breaking_changes": [f"compare:{old_spec_path}->{new_spec_path}"],
        "deprecations": [],
        "new_features": [],
    }
