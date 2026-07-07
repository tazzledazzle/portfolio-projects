from changelog_generator.transforms.python_rewriter import generate_python_migration


def test_generate_python_migration_noop() -> None:
    result = generate_python_migration({"breaking_changes": []})
    assert result == "# no-op"
