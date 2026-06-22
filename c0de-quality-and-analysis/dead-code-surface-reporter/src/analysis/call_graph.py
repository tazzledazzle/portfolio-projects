def find_unreachable_symbols(nodes: list[dict]) -> list[dict]:
    return [node for node in nodes if not node.get("reachable", False)]
