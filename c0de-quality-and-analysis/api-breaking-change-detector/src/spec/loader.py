import json


def load_openapi(path: str) -> dict:
    # JSON loader placeholder; can be replaced with yaml parser.
    with open(path, "r", encoding="utf-8") as handle:
        return json.load(handle)
