from pathlib import Path
import json
import hashlib

import pytest

from ai_code_assistant.cli import main


def test_cli_requires_exactly_one_target(tmp_path: Path) -> None:
    file_path = tmp_path / "module.py"
    file_path.write_text("def f():\n    return 1\n", encoding="utf-8")

    with pytest.raises(SystemExit):
        main(["gen-tests"])

    with pytest.raises(SystemExit):
        main(["gen-tests", str(file_path), "--repo", str(tmp_path)])


def test_cli_single_file_dry_run(tmp_path: Path, capsys: pytest.CaptureFixture[str]) -> None:
    source = tmp_path / "module.py"
    source.write_text("def f():\n    return 1\n", encoding="utf-8")

    code = main(["gen-tests", str(source), "--dry-run"])
    output = capsys.readouterr().out

    assert code == 0
    assert "test_module.py" in output
    assert "def test_f_behavior" in output
    assert "assert result == 1" in output
    assert not (tmp_path / "tests" / "unit" / "test_module.py").exists()


def test_cli_repo_mode_writes_tests(tmp_path: Path) -> None:
    app = tmp_path / "app"
    app.mkdir()
    (app / "a.py").write_text("def a():\n    return 1\n", encoding="utf-8")
    (app / "b.py").write_text("def b():\n    return 2\n", encoding="utf-8")

    code = main(["gen-tests", "--repo", str(tmp_path)])

    assert code == 0
    assert (tmp_path / "tests" / "unit" / "test_a.py").exists()
    assert (tmp_path / "tests" / "unit" / "test_b.py").exists()


def test_cli_read_only_profile_requires_dry_run(tmp_path: Path) -> None:
    source = tmp_path / "module.py"
    source.write_text("def f():\n    return 1\n", encoding="utf-8")

    with pytest.raises(SystemExit):
        main(["gen-tests", str(source), "--profile", "read-only"])


def test_cli_single_file_json_output(tmp_path: Path, capsys: pytest.CaptureFixture[str]) -> None:
    source = tmp_path / "module.py"
    source.write_text("def f():\n    return 1\n", encoding="utf-8")

    code = main(["gen-tests", str(source), "--dry-run", "--output", "json"])
    payload = json.loads(capsys.readouterr().out)

    assert code == 0
    assert payload["status"] == "ok"
    assert payload["dry_run"] is True
    assert payload["processed_count"] == 1
    assert payload["results"][0]["action"] == "previewed"
    assert payload["results"][0]["target_path"].endswith("test_module.py")


def test_cli_repo_json_output(tmp_path: Path, capsys: pytest.CaptureFixture[str]) -> None:
    app = tmp_path / "app"
    app.mkdir()
    (app / "a.py").write_text("def a():\n    return 1\n", encoding="utf-8")
    (app / "b.py").write_text("def b():\n    return 2\n", encoding="utf-8")

    code = main(["gen-tests", "--repo", str(tmp_path), "--dry-run", "--output", "json"])
    payload = json.loads(capsys.readouterr().out)

    assert code == 0
    assert payload["status"] == "ok"
    assert payload["dry_run"] is True
    assert payload["processed_count"] == 2
    assert len(payload["results"]) == 2


def test_cli_pyramid_all_generates_three_levels(tmp_path: Path, capsys: pytest.CaptureFixture[str]) -> None:
    source = tmp_path / "module.py"
    source.write_text("def f():\n    return 1\n", encoding="utf-8")

    code = main(["gen-tests", str(source), "--dry-run", "--output", "json", "--pyramid", "all"])
    payload = json.loads(capsys.readouterr().out)

    assert code == 0
    assert payload["processed_count"] == 3
    targets = [result["target_path"] for result in payload["results"]]
    assert any("/tests/unit/" in path for path in targets)
    assert any("/tests/integration/" in path for path in targets)
    assert any("/tests/e2e/" in path for path in targets)


def test_cli_writes_audit_log_entries(tmp_path: Path) -> None:
    source = tmp_path / "module.py"
    source.write_text("def f():\n    return 1\n", encoding="utf-8")
    audit_log = tmp_path / ".ai-code-assistant" / "audit.log.jsonl"

    code = main(["gen-tests", str(source), "--dry-run", "--audit-log", str(audit_log)])

    assert code == 0
    lines = [json.loads(line) for line in audit_log.read_text(encoding="utf-8").splitlines()]
    assert len(lines) == 2
    assert lines[0]["event_type"] == "file_result"
    assert lines[0]["action"] == "previewed"
    assert lines[1]["event_type"] == "run_summary"
    assert lines[1]["processed_count"] == 1


def test_extensions_validate_manifest_json(tmp_path: Path, capsys: pytest.CaptureFixture[str]) -> None:
    core = {
        "name": "demo",
        "version": "0.1.0",
        "entrypoint": "demo.py",
        "capabilities": ["read_repo"],
    }
    sha = hashlib.sha256(json.dumps(core, sort_keys=True).encode("utf-8")).hexdigest()
    manifest_path = tmp_path / "manifest.json"
    manifest_path.write_text(json.dumps({**core, "sha256": sha, "signature": "x"}), encoding="utf-8")

    code = main(["extensions", "validate-manifest", "--manifest", str(manifest_path), "--output", "json"])
    payload = json.loads(capsys.readouterr().out)
    assert code == 0
    assert payload["status"] == "ok"


def test_run_plan_json_output(tmp_path: Path, capsys: pytest.CaptureFixture[str]) -> None:
    plan_path = tmp_path / "plan.json"
    plan_path.write_text(
        json.dumps({"steps": [{"id": "s1", "command": "echo ok", "verify_command": "echo verify"}]}),
        encoding="utf-8",
    )
    code = main(["run-plan", "--plan", str(plan_path), "--output", "json"])
    payload = json.loads(capsys.readouterr().out)
    assert code == 0
    assert payload["status"] == "ok"
    assert payload["steps"][0]["id"] == "s1"
