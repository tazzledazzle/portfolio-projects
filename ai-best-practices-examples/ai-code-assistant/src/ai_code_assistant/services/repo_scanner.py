from pathlib import Path

EXCLUDED_DIRS = {
    ".git",
    ".hg",
    ".svn",
    ".venv",
    "venv",
    "__pycache__",
    ".mypy_cache",
    ".pytest_cache",
    "node_modules",
    "tests",
}


def scan_python_files(repo_path: str | Path) -> list[Path]:
    root = Path(repo_path).resolve()
    files: list[Path] = []
    for path in root.rglob("*.py"):
        if _is_excluded(path=path, root=root):
            continue
        files.append(path)
    return sorted(files)


def _is_excluded(path: Path, root: Path) -> bool:
    rel = path.resolve().relative_to(root)
    parts = rel.parts[:-1]
    for part in parts:
        if part.startswith(".") or part in EXCLUDED_DIRS:
            return True
    if path.name.startswith("test_") or path.name.endswith("_test.py"):
        return True
    return False
