from pathlib import Path
from typing import Any

import yaml


def load_tool_versions(path: Path) -> dict[str, str]:
    payload: Any = yaml.safe_load(path.read_text())
    if not isinstance(payload, dict):
        raise ValueError(f"Manifest must be a mapping: {path}")
    tools = payload.get("tools", payload)
    if not isinstance(tools, dict):
        raise ValueError(f"Manifest tools section must be a mapping: {path}")
    return {str(key): str(value) for key, value in tools.items()}


def load_canonical_policy(path: Path) -> dict[str, str]:
    return load_tool_versions(path)
