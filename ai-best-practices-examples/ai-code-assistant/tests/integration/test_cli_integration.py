from pathlib import Path

import pytest

from ai_code_assistant import cli


def test_single_file_mode_calls_output_with_expected_target(
    tmp_path: Path, monkeypatch: pytest.MonkeyPatch
) -> None:
    monkeypatch.chdir(tmp_path)
    source = tmp_path / "engine.py"
    source.write_text("def run():\n    return 1\n", encoding="utf-8")

    captured: list[tuple[Path, str, bool, str, Path, Path]] = []

    def fake_output(
        target_path: Path,
        content: str,
        dry_run: bool,
        output_mode: str,
        audit_log_path: Path,
        source_path: Path,
        profile: str,
        policy,
        approve_high_risk: bool,
    ) -> dict[str, str | bool]:
        del profile, policy, approve_high_risk
        captured.append((target_path, content, dry_run, output_mode, audit_log_path, source_path))
        return {"target_path": str(target_path), "action": "previewed", "content": content}

    monkeypatch.setattr(cli, "_output_one", fake_output)

    exit_code = cli.main(["gen-tests", str(source), "--dry-run"])

    assert exit_code == 0
    assert len(captured) == 1
    target_path, content, dry_run, output_mode, _audit_log_path, source_path = captured[0]
    assert target_path == tmp_path / "tests" / "unit" / "test_engine.py"
    assert "def test_run_behavior" in content
    assert "assert result == 1" in content
    assert dry_run is True
    assert output_mode == "text"
    assert source_path == source


def test_repo_mode_processes_all_scanned_files(tmp_path: Path, monkeypatch: pytest.MonkeyPatch) -> None:
    (tmp_path / "pkg").mkdir()
    (tmp_path / "pkg" / "alpha.py").write_text("def alpha():\n    return 1\n", encoding="utf-8")
    (tmp_path / "pkg" / "beta.py").write_text("def beta():\n    return 2\n", encoding="utf-8")
    (tmp_path / "pkg" / "test_beta.py").write_text("def test_beta():\n    pass\n", encoding="utf-8")

    written: list[Path] = []

    def fake_output(
        target_path: Path,
        content: str,
        dry_run: bool,
        output_mode: str,
        audit_log_path: Path,
        source_path: Path,
        profile: str,
        policy,
        approve_high_risk: bool,
    ) -> dict[str, str | bool]:
        del content, dry_run, audit_log_path, source_path, profile, policy, approve_high_risk
        assert output_mode == "text"
        written.append(target_path)
        return {"target_path": str(target_path), "action": "wrote", "content": ""}

    monkeypatch.setattr(cli, "_output_one", fake_output)

    exit_code = cli.main(["gen-tests", "--repo", str(tmp_path)])

    assert exit_code == 0
    assert written == [
        tmp_path / "tests" / "unit" / "test_alpha.py",
        tmp_path / "tests" / "unit" / "test_beta.py",
    ]
