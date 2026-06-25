from ai_code_assistant.adapters.llm_adapter import LLMAdapter


def test_adapter_uses_fallback_without_api_key() -> None:
    adapter = LLMAdapter(api_key=None)
    content = adapter.generate_tests(source_code="def x():\n    return 1\n", module_name="sample")

    assert "def test_x_behavior" in content
    assert "assert result == 1" in content
    assert "assert True" not in content


def test_adapter_uses_openai_path_when_key_present(monkeypatch) -> None:
    adapter = LLMAdapter(api_key="abc123")

    def fake_generate_with_openai(
        source_code: str,
        module_name: str,
        facts,
        test_level: str,
        source_path=None,
        repo_root=None,
    ) -> str:
        del source_code, module_name, facts, test_level, source_path, repo_root
        return "def test_from_model():\n    assert 1 == 1\n"

    monkeypatch.setattr(adapter, "_generate_with_openai", fake_generate_with_openai)
    content = adapter.generate_tests(source_code="def x():\n    return 1\n", module_name="sample")

    assert "def test_from_model" in content
