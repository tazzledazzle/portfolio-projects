"""Load pipeline-failure fixtures and summarize them with an offline LLM."""

from __future__ import annotations

import json
from pathlib import Path
from typing import Any

from llm_fixture import OfflineFixtureLLM

SECRET_FIELDS = {"api_key", "authorization", "password", "secret", "token"}


def _redact_secret_fields(value: Any) -> Any:
    if isinstance(value, dict):
        return {
            key: "[REDACTED]" if key.lower() in SECRET_FIELDS else _redact_secret_fields(item)
            for key, item in value.items()
        }
    if isinstance(value, list):
        return [_redact_secret_fields(item) for item in value]
    return value


def _normalize_failures(payload: Any) -> list[dict]:
    failures = payload.get("failures") if isinstance(payload, dict) else payload
    if not isinstance(failures, list) or not failures:
        raise ValueError("failures must be a non-empty list")
    normalized = []
    for index, failure in enumerate(failures):
        if not isinstance(failure, dict):
            raise ValueError(f"failure at index {index} must be an object")
        prompt_key = failure.get("id")
        summary = failure.get("summary")
        if not isinstance(prompt_key, str) or not prompt_key.strip():
            raise ValueError(f"failure at index {index} requires an id")
        if not isinstance(summary, str) or not summary.strip():
            raise ValueError(f"failure {prompt_key!r} requires a summary")
        normalized.append(_redact_secret_fields(failure))
    return normalized


def load_failure_fixtures(path: str | Path) -> list[dict]:
    with Path(path).open(encoding="utf-8") as fixture_file:
        return _normalize_failures(json.load(fixture_file))


def summarize_failures(failures: Any) -> dict:
    normalized = _normalize_failures(failures)
    fixtures = {failure["id"]: failure for failure in normalized}
    llm = OfflineFixtureLLM(fixtures)
    completions = [llm.complete(failure["id"]) for failure in normalized]
    findings = [
        finding
        for completion in completions
        for finding in completion["findings"]
    ]
    return {
        "summary": " ".join(completion["text"] for completion in completions),
        "findings": findings,
        "failures_ingested": len(completions),
        "offline_llm": True,
        "offline_fixture_llm": True,
        "fixture_mode": True,
        "live_provider": False,
        "provider": llm.provider,
    }
