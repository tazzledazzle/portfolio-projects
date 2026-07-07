import os
import subprocess
import sys
from pathlib import Path


def _run_cli(args: list[str], cwd: Path) -> subprocess.CompletedProcess[str]:
    env = os.environ.copy()
    project_src = Path(__file__).resolve().parents[2] / "src"
    env["PYTHONPATH"] = str(project_src)
    return subprocess.run(
        [sys.executable, "-m", "ai_code_assistant.cli", *args],
        cwd=cwd,
        env=env,
        check=False,
        text=True,
        capture_output=True,
    )


def test_e2e_single_file_dry_run_prints_generated_output(tmp_path: Path) -> None:
    src_dir = tmp_path / "src"
    src_dir.mkdir()
    (src_dir / "worker.py").write_text("def work():\n    return 1\n", encoding="utf-8")

    result = _run_cli(["gen-tests", "src/worker.py", "--dry-run"], cwd=tmp_path)

    assert result.returncode == 0
    assert "---" in result.stdout
    assert "test_worker.py" in result.stdout
    assert "def test_work_behavior" in result.stdout
    assert "assert result == 1" in result.stdout


def test_e2e_repo_mode_writes_test_files(tmp_path: Path) -> None:
    pkg = tmp_path / "pkg"
    pkg.mkdir()
    (pkg / "service.py").write_text("def service():\n    return 'ok'\n", encoding="utf-8")

    result = _run_cli(["gen-tests", "--repo", "."], cwd=tmp_path)

    assert result.returncode == 0
    generated = tmp_path / "tests" / "unit" / "test_service.py"
    assert generated.exists()
    content = generated.read_text(encoding="utf-8")
    assert "def test_service_behavior" in content
    assert "assert result == 'ok'" in content
