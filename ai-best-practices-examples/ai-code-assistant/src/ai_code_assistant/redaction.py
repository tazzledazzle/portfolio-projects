import re


SENSITIVE_KEYS = {"api_key", "token", "secret", "password", "authorization"}


def redact_text(value: str, patterns: list[str], replacement: str = "***REDACTED***") -> str:
    redacted = value
    for pattern in patterns:
        redacted = re.sub(pattern, replacement, redacted)
    return redacted


def redact_mapping(data: dict, patterns: list[str], replacement: str = "***REDACTED***") -> dict:
    out = {}
    for key, value in data.items():
        key_lower = str(key).lower()
        if isinstance(value, dict):
            out[key] = redact_mapping(value, patterns, replacement)
            continue
        if key_lower in SENSITIVE_KEYS:
            out[key] = replacement
            continue
        if isinstance(value, str):
            out[key] = redact_text(value, patterns, replacement)
        else:
            out[key] = value
    return out
