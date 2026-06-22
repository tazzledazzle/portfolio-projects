from pathlib import Path

from ai_code_assistant.services.repo_scanner import scan_python_files


def test_scanner_excludes_common_noise_dirs(tmp_path: Path) -> None:
    (tmp_path / "src").mkdir()
    (tmp_path / "src" / "main.py").write_text("x = 1\n", encoding="utf-8")
    (tmp_path / "tests").mkdir()
    (tmp_path / "tests" / "test_main.py").write_text("def test_x():\n    pass\n", encoding="utf-8")
    (tmp_path / "__pycache__").mkdir()
    (tmp_path / "__pycache__" / "cache.py").write_text("x = 2\n", encoding="utf-8")
    (tmp_path / ".pytest_cache").mkdir()
    (tmp_path / ".pytest_cache" / "cache.py").write_text("x = 3\n", encoding="utf-8")

    files = scan_python_files(tmp_path)

    assert files == [tmp_path / "src" / "main.py"]


def test_scanner_excludes_test_named_files_anywhere(tmp_path: Path) -> None:
    (tmp_path / "pkg").mkdir()
    (tmp_path / "pkg" / "service.py").write_text("x = 1\n", encoding="utf-8")
    (tmp_path / "pkg" / "test_service.py").write_text("def test_x():\n    pass\n", encoding="utf-8")
    (tmp_path / "pkg" / "service_test.py").write_text("def test_x():\n    pass\n", encoding="utf-8")

    files = scan_python_files(tmp_path)

    assert files == [tmp_path / "pkg" / "service.py"]


def test_scanner_returns_empty_when_no_candidates(tmp_path: Path) -> None:
    (tmp_path / "README.md").write_text("hello\n", encoding="utf-8")

    assert scan_python_files(tmp_path) == []
