from pathlib import Path

from summarizer import load_failure_fixtures, summarize_failures


FIXTURES = Path(__file__).parent / "testdata" / "failures.json"


def test_summarize_failures_ingests_fixtures():
    fixtures = load_failure_fixtures(FIXTURES)

    result = summarize_failures(fixtures)

    assert fixtures[0]["api_key"] == "[REDACTED]"
    assert result["failures_ingested"] >= 1
    assert result["summary"]
    assert result["offline_llm"] is True
    assert result["fixture_mode"] is True
    assert result["live_provider"] is False


def test_summarize_never_requires_api_keys(monkeypatch):
    monkeypatch.delenv("OPENAI_API_KEY", raising=False)
    monkeypatch.delenv("ANTHROPIC_API_KEY", raising=False)

    result = summarize_failures(load_failure_fixtures(FIXTURES))

    assert result["summary"]
    assert result["live_provider"] is False
