def generate_kotlin_migration(delta: dict[str, list[str]]) -> str:
    if not delta.get("breaking_changes"):
        return "// no-op"
    return "// apply migration patches for breaking API changes"
