"""Deterministic, offline substitute for an LLM completion provider."""

from __future__ import annotations

from copy import deepcopy


class OfflineFixtureLLM:
    """Return pre-authored completions without network access or API keys."""

    def __init__(self, fixtures: dict[str, dict]):
        self.fixtures = deepcopy(fixtures)
        self.provider = "offline_fixture"
        self.live = False

    def complete(self, prompt_key: str) -> dict:
        if prompt_key not in self.fixtures:
            raise KeyError(prompt_key)
        fixture = self.fixtures[prompt_key]
        summary = fixture.get("summary")
        if not isinstance(summary, str) or not summary.strip():
            raise ValueError(f"fixture {prompt_key!r} requires a non-empty summary")
        findings = fixture.get("findings", [])
        if not isinstance(findings, list):
            raise ValueError(f"fixture {prompt_key!r} findings must be a list")
        return {
            "text": summary.strip(),
            "findings": list(findings),
            "offline_llm": True,
            "live_provider": False,
        }
