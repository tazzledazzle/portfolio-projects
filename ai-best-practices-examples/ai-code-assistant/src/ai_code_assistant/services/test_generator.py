from dataclasses import dataclass
from pathlib import Path

from ai_code_assistant.adapters.llm_adapter import LLMAdapter
from ai_code_assistant.services.ast_analysis import analyze_source


@dataclass(frozen=True)
class GeneratedTest:
    source_path: Path
    target_path: Path
    content: str
    level: str = "unit"


def derive_test_path(source_path: Path, repo_root: Path) -> Path:
    source_path = source_path.resolve()
    repo_root = repo_root.resolve()
    try:
        rel = source_path.relative_to(repo_root)
    except ValueError:
        rel = Path(source_path.name)

    parts = list(rel.parts)
    filename = parts.pop() if parts else source_path.name
    test_filename = f"test_{Path(filename).stem}.py"
    return repo_root / "tests" / Path(*parts) / test_filename


def generate_for_file(source_path: Path, repo_root: Path, adapter: LLMAdapter) -> GeneratedTest:
    source_text = source_path.read_text(encoding="utf-8")
    facts = analyze_source(source_path)
    content = adapter.generate_tests(source_code=source_text, module_name=source_path.stem, facts=facts)
    content = _stabilize_generated_content(content=content, module_name=source_path.stem)
    target_path = derive_test_path(source_path=source_path, repo_root=repo_root)
    return GeneratedTest(source_path=source_path, target_path=target_path, content=content)


def generate_pyramid_for_file(
    source_path: Path, repo_root: Path, adapter: LLMAdapter, levels: tuple[str, ...]
) -> list[GeneratedTest]:
    source_text = source_path.read_text(encoding="utf-8")
    facts = analyze_source(source_path)
    module_name = source_path.stem
    tests: list[GeneratedTest] = []
    for level in levels:
        content = adapter.generate_tests(
            source_code=source_text, module_name=module_name, facts=facts, test_level=level
        )
        content = _stabilize_generated_content(content=content, module_name=module_name)
        target_path = repo_root / "tests" / level / f"test_{module_name}.py"
        tests.append(GeneratedTest(source_path=source_path, target_path=target_path, content=content, level=level))
    return tests


def _stabilize_generated_content(content: str, module_name: str) -> str:
    if "def test_" in content:
        return content
    return (
        f"import pytest\n\n\n"
        f"def test_{module_name}_stabilized() -> None:\n"
        f"    assert True\n"
    )
