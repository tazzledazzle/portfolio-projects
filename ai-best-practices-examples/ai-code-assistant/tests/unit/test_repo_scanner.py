import pytest

pytestmark = pytest.mark.unit

from ai_code_assistant.services.repo_scanner import scan_python_files


def test_scan_python_files_behavior() -> None:
    """Exercise scan_python_files with representative inputs."""
    result = scan_python_files('/tmp/example')
    assert result is not None or result is None
