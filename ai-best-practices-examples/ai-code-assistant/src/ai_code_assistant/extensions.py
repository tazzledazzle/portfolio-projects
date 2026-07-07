import hashlib
import json
from pathlib import Path

REQUIRED_MANIFEST_KEYS = {
    "name",
    "version",
    "entrypoint",
    "capabilities",
    "sha256",
    "signature",
}


def validate_manifest(manifest_path: Path) -> dict:
    payload = json.loads(manifest_path.read_text(encoding="utf-8"))
    missing = REQUIRED_MANIFEST_KEYS - set(payload.keys())
    if missing:
        raise ValueError(f"Missing required manifest fields: {sorted(missing)}")
    if not isinstance(payload["capabilities"], list):
        raise ValueError("capabilities must be a list.")
    digest = hashlib.sha256(
        json.dumps(
            {
                "name": payload["name"],
                "version": payload["version"],
                "entrypoint": payload["entrypoint"],
                "capabilities": payload["capabilities"],
            },
            sort_keys=True,
        ).encode("utf-8")
    ).hexdigest()
    if digest != payload["sha256"]:
        raise ValueError("sha256 checksum does not match manifest payload.")
    return payload
