import json


def write_sbom(packages: list[dict], path: str) -> None:
    document = {
        "bomFormat": "CycloneDX",
        "specVersion": "1.5",
        "components": packages,
    }
    with open(path, "w", encoding="utf-8") as handle:
        json.dump(document, handle, indent=2)
