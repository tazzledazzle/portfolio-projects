from ai_code_assistant.adapters.llm_adapter import LLMAdapter


def test_fallback_contains_pytest_and_placeholder_name() -> None:
    adapter = LLMAdapter(api_key=None)

    content = adapter.generate_tests(source_code="def a():\n    return 1\n", module_name="payments")

    assert "import pytest" in content
    assert "def test_payments_unit_placeholder" in content


def test_generate_prefers_openai_path_when_api_key_is_set(monkeypatch) -> None:
    adapter = LLMAdapter(api_key="token")

    def fake_openai(source_code: str, module_name: str, facts, test_level: str) -> str:
        assert "return 1" in source_code
        assert module_name == "calc"
        assert test_level == "unit"
        del facts
        return "def test_calc_model():\n    assert True\n"

    monkeypatch.setattr(adapter, "_generate_with_openai", fake_openai)

    content = adapter.generate_tests(source_code="def add():\n    return 1\n", module_name="calc")

    assert "def test_calc_model" in content


def test_openai_path_falls_back_if_empty_response(monkeypatch) -> None:
    adapter = LLMAdapter(api_key="token")

    def fake_openai(source_code: str, module_name: str, facts, test_level: str) -> str:
        del source_code, module_name, facts, test_level
        return ""

    monkeypatch.setattr(adapter, "_generate_with_openai", fake_openai)

    content = adapter.generate_tests(source_code="def add():\n    return 1\n", module_name="calc")

    assert "def test_calc_unit_placeholder" in content
