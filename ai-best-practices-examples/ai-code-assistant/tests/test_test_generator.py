from pathlib import Path

from ai_code_assistant.adapters.llm_adapter import LLMAdapter
from ai_code_assistant.services.test_generator import derive_test_path, generate_for_file


def test_derive_test_path_for_nested_source(tmp_path: Path) -> None:
    src = tmp_path / "pkg" / "utils.py"
    src.parent.mkdir(parents=True)
    src.write_text("def f():\n    return 1\n", encoding="utf-8")

    target = derive_test_path(source_path=src, repo_root=tmp_path)

    assert target == tmp_path / "tests" / "pkg" / "test_utils.py"


def test_generate_for_file_uses_adapter_and_returns_pytest_style(tmp_path: Path) -> None:
    src = tmp_path / "calc.py"
    src.write_text("def add(a, b):\n    return a + b\n", encoding="utf-8")

    result = generate_for_file(source_path=src, repo_root=tmp_path, adapter=LLMAdapter(api_key=None))

    assert result.target_path == tmp_path / "tests" / "test_calc.py"
    assert "def test_" in result.content
    assert "pytest" in result.content
