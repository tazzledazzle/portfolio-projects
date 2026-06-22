from pathlib import Path

from ai_code_assistant.services.test_generator import derive_test_path, generate_for_file


class StubAdapter:
    def __init__(self, content: str) -> None:
        self._content = content
        self.calls: list[tuple[str, str]] = []

    def generate_tests(self, source_code: str, module_name: str, facts=None, test_level: str = "unit") -> str:
        del facts, test_level
        self.calls.append((source_code, module_name))
        return self._content


def test_derive_path_keeps_relative_structure(tmp_path: Path) -> None:
    src_file = tmp_path / "src" / "pkg" / "mod.py"
    src_file.parent.mkdir(parents=True)
    src_file.write_text("def f():\n    return 1\n", encoding="utf-8")

    test_path = derive_test_path(src_file, tmp_path)

    assert test_path == tmp_path / "tests" / "src" / "pkg" / "test_mod.py"


def test_derive_path_outside_repo_falls_back_to_filename(tmp_path: Path) -> None:
    external = tmp_path.parent / "external_mod.py"
    external.write_text("x = 1\n", encoding="utf-8")
    try:
        test_path = derive_test_path(external, tmp_path)
        assert test_path == tmp_path / "tests" / "test_external_mod.py"
    finally:
        if external.exists():
            external.unlink()


def test_generate_for_file_passes_source_and_module_name(tmp_path: Path) -> None:
    source = tmp_path / "billing.py"
    source_code = "def total(a, b):\n    return a + b\n"
    source.write_text(source_code, encoding="utf-8")
    adapter = StubAdapter("def test_total():\n    assert True\n")

    result = generate_for_file(source_path=source, repo_root=tmp_path, adapter=adapter)

    assert result.source_path == source
    assert result.target_path == tmp_path / "tests" / "test_billing.py"
    assert result.content == "def test_total():\n    assert True\n"
    assert adapter.calls == [(source_code, "billing")]
