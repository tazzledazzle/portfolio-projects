import os

from ai_code_assistant.services.ast_analysis import AnalysisFacts

class LLMAdapter:
    """Generates unit tests with OpenAI when configured, otherwise fallback."""

    def __init__(self, api_key: str | None = None) -> None:
        self._api_key = api_key if api_key is not None else os.getenv("OPENAI_API_KEY")

    def generate_tests(
        self,
        source_code: str,
        module_name: str,
        facts: AnalysisFacts | None = None,
        test_level: str = "unit",
    ) -> str:
        if self._api_key:
            response = self._generate_with_openai(
                source_code=source_code,
                module_name=module_name,
                facts=facts,
                test_level=test_level,
            )
            if response.strip():
                return response
            return self._fallback_template(module_name=module_name, facts=facts, test_level=test_level)
        return self._fallback_template(module_name=module_name, facts=facts, test_level=test_level)

    def _generate_with_openai(
        self,
        source_code: str,
        module_name: str,
        facts: AnalysisFacts | None,
        test_level: str,
    ) -> str:
        try:
            from openai import OpenAI  # type: ignore
        except Exception:
            return self._fallback_template(module_name=module_name)

        client = OpenAI(api_key=self._api_key)
        prompt = (
            f"Write concise pytest {test_level} tests for this Python module. "
            "Only return Python code.\n\n"
            f"Module name: {module_name}\n\n"
            f"Functions: {facts.function_names if facts else []}\n"
            f"Classes: {facts.class_names if facts else []}\n"
            f"{source_code}"
        )
        response = client.responses.create(
            model="gpt-4.1-mini",
            input=prompt,
            temperature=0,
        )
        text = getattr(response, "output_text", "").strip()
        if not text:
            return self._fallback_template(module_name=module_name, facts=facts, test_level=test_level)
        return text

    @staticmethod
    def _fallback_template(module_name: str, facts: AnalysisFacts | None, test_level: str) -> str:
        marker = test_level
        target_name = facts.function_names[0] if facts and facts.function_names else module_name
        return (
            f"import pytest\n\n\n"
            f"pytestmark = pytest.mark.{marker}\n\n\n"
            f"def test_{target_name}_{test_level}_placeholder() -> None:\n"
            f"    \"\"\"Replace this placeholder with real test cases.\"\"\"\n"
            f"    assert True\n"
        )
