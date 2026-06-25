from pathlib import Path

from ai_code_assistant.services.test_codegen import derive_module_import, generate_robust_tests


def test_derive_module_import_strips_src_layout() -> None:
    repo = Path("/repo")
    source = Path("/repo/src/pkg/mod.py")

    assert derive_module_import(source, repo, "mod") == "pkg.mod"


def test_generate_unit_tests_call_function_and_assert_return() -> None:
    source = "def add(a, b):\n    return a + b\n"

    content = generate_robust_tests(
        source_code=source,
        module_name="calc",
        facts=None,
        test_level="unit",
    )

    assert "from calc import add" in content
    assert "def test_add_behavior" in content
    assert "add(2, 3)" in content
    assert "assert result == 5" in content
    assert "pytest.mark.unit" in content
    assert "assert True" not in content


def test_generate_integration_tests_multiple_functions() -> None:
    source = "def alpha():\n    return 1\n\ndef beta():\n    return 2\n"

    content = generate_robust_tests(
        source_code=source,
        module_name="workflow",
        facts=None,
        test_level="integration",
    )

    assert "def test_workflow_functions_work_together" in content
    assert "alpha()" in content
    assert "beta()" in content


def test_generate_e2e_tests_main_entrypoint() -> None:
    source = "def main():\n    return 0\n"

    content = generate_robust_tests(
        source_code=source,
        module_name="runner",
        facts=None,
        test_level="e2e",
    )

    assert "def test_runner_main_entrypoint" in content
    assert "main()" in content
    assert "assert result == 0" in content
