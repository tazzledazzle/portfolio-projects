def normalize_license(raw_license: str) -> str:
    normalized = raw_license.strip().replace(" ", "-")
    aliases = {"Apache 2.0": "Apache-2.0", "MIT License": "MIT"}
    return aliases.get(normalized, normalized or "UNKNOWN")
