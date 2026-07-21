import pytest

from llm_fixture import OfflineFixtureLLM


def test_offline_fixture_llm_complete_returns_summary():
    llm = OfflineFixtureLLM(
        {
            "pipeline-42": {
                "summary": "Tests failed because the cache service was unavailable.",
                "findings": ["cache connection refused"],
            }
        }
    )

    result = llm.complete("pipeline-42")

    assert result["text"]
    assert result["findings"] == ["cache connection refused"]
    assert result["offline_llm"] is True
    assert result["live_provider"] is False
    assert llm.provider == "offline_fixture"
    assert llm.live is False


def test_offline_fixture_llm_unknown_key_raises():
    llm = OfflineFixtureLLM({})

    with pytest.raises(KeyError):
        llm.complete("missing")
