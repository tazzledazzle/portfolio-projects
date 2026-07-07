from pathlib import Path

from ai_code_assistant.services.repo_scanner import scan_python_files


def test_scan_python_files_excludes_ignored_dirs(tmp_path: Path) -> None:
    (tmp_path / "app").mkdir()
    (tmp_path / "app" / "service.py").write_text("def run():\n    return 1\n", encoding="utf-8")

    (tmp_path / "tests").mkdir()
    (tmp_path / "tests" / "test_service.py").write_text("def test_x():\n    assert True\n", encoding="utf-8")

    (tmp_path / ".venv").mkdir()
    (tmp_path / ".venv" / "ignored.py").write_text("x=1\n", encoding="utf-8")

    (tmp_path / ".hidden").mkdir()
    (tmp_path / ".hidden" / "hidden.py").write_text("x=2\n", encoding="utf-8")

    found = scan_python_files(tmp_path)

    assert found == [tmp_path / "app" / "service.py"]


def test_scan_python_files_returns_sorted_paths(tmp_path: Path) -> None:
    (tmp_path / "b.py").write_text("b=1\n", encoding="utf-8")
    (tmp_path / "a.py").write_text("a=1\n", encoding="utf-8")
    (tmp_path / "test_a.py").write_text("def test_x():\n    assert True\n", encoding="utf-8")

    found = scan_python_files(tmp_path)

    assert found == [tmp_path / "a.py", tmp_path / "b.py"]
