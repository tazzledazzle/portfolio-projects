def parse_gradle_tree(raw_text: str) -> list[dict]:
    """Convert Gradle dependency output into package records."""
    packages: list[dict] = []
    for line in raw_text.splitlines():
        if ":" not in line:
            continue
        token = line.strip().split(" ")[0]
        if token.count(":") >= 2:
            packages.append({"name": token, "license": "UNKNOWN"})
    return packages
