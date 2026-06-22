import hashlib
import json
from pathlib import Path

from ai_code_assistant.extensions import validate_manifest


def test_validate_manifest_happy_path(tmp_path: Path) -> None:
    core = {
        "name": "demo",
        "version": "0.1.0",
        "entrypoint": "demo.py",
        "capabilities": ["read_repo"],
    }
    sha = hashlib.sha256(json.dumps(core, sort_keys=True).encode("utf-8")).hexdigest()
    payload = {**core, "sha256": sha, "signature": "placeholder"}
    manifest = tmp_path / "extension-manifest.v1.json"
    manifest.write_text(json.dumps(payload), encoding="utf-8")

    loaded = validate_manifest(manifest)
    assert loaded["name"] == "demo"
