def diff_paths(base_paths: dict, head_paths: dict) -> list[dict]:
    changes: list[dict] = []
    removed = sorted(set(base_paths.keys()) - set(head_paths.keys()))
    for path in removed:
        changes.append({"kind": "path_removed", "path": path})
    return changes
