import os
from pathlib import Path

from ai_code_assistant.services.ast_analysis import AnalysisFacts
from ai_code_assistant.services.test_codegen import generate_robust_tests


class LLMAdapter:
    """Generates pytest tests with OpenAI when configured, otherwise AST-driven codegen."""

    def __init__(self, api_key: str | None = None) -> None:
        self._api_key = api_key if api_key is not None else os.getenv("OPENAI_API_KEY")

    def generate_tests(
        self,
        source_code: str,
        module_name: str,
        facts: AnalysisFacts | None = None,
        test_level: str = "unit",
        source_path: Path | None = None,
        repo_root: Path | None = None,
    ) -> str:
        if self._api_key:
            response = self._generate_with_openai(
                source_code=source_code,
                module_name=module_name,
                facts=facts,
                test_level=test_level,
                source_path=source_path,
                repo_root=repo_root,
            )
            if response.strip():
                return response
        return generate_robust_tests(
            source_code=source_code,
            module_name=module_name,
            facts=facts,
            test_level=test_level,
            source_path=source_path,
            repo_root=repo_root,
        )

    def _generate_with_openai(
        self,
        source_code: str,
        module_name: str,
        facts: AnalysisFacts | None,
        test_level: str,
        source_path: Path | None,
        repo_root: Path | None,
    ) -> str:
        try:
            from openai import OpenAI  # type: ignore
        except Exception:
            return generate_robust_tests(
                source_code=source_code,
                module_name=module_name,
                facts=facts,
                test_level=test_level,
                source_path=source_path,
                repo_root=repo_root,
            )

        from ai_code_assistant.services.test_codegen import derive_module_import

        module_import = derive_module_import(source_path, repo_root, module_name)
        level_guidance = {
            "unit": (
                "Write fast isolated tests. Mock external I/O (os, requests, subprocess) with unittest.mock. "
                "Import and call real functions from the module."
            ),
            "integration": (
                "Write tests that exercise multiple functions or module wiring together. "
                "Use tmp_path for filesystem interactions when needed."
            ),
            "e2e": (
                "Write an end-to-end test that exercises the module's primary workflow or main() entrypoint. "
                "Use subprocess only when the module is intended to run as a script."
            ),
        }.get(test_level, "Write focused pytest tests.")

        prompt = (
            f"Write pytest {test_level} tests for this Python module.\n"
            "Requirements:\n"
            "- Return only valid Python test code (no markdown fences).\n"
            f"- Import from `{module_import}` using explicit imports.\n"
            "- Every test must call real code from the module and assert on behavior.\n"
            "- Do not use `assert True` or empty placeholder tests.\n"
            f"- Tag the module with `pytestmark = pytest.mark.{test_level}`.\n"
            f"- {level_guidance}\n\n"
            f"Module name: {module_name}\n"
            f"Import path: {module_import}\n"
            f"Functions: {facts.function_names if facts else []}\n"
            f"Classes: {facts.class_names if facts else []}\n"
            f"Imports used: {facts.import_names if facts else []}\n"
            f"Has async functions: {facts.has_async if facts else False}\n\n"
            f"{source_code}"
        )
        client = OpenAI(api_key=self._api_key)
        response = client.responses.create(
            model="gpt-4.1-mini",
            input=prompt,
            temperature=0,
        )
        text = getattr(response, "output_text", "").strip()
        if not text:
            return generate_robust_tests(
                source_code=source_code,
                module_name=module_name,
                facts=facts,
                test_level=test_level,
                source_path=source_path,
                repo_root=repo_root,
            )
        return text
