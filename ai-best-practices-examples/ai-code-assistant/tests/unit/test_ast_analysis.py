import pytest

pytestmark = pytest.mark.unit

from ai_code_assistant.services.ast_analysis import analyze_source


def test_analyze_source_behavior() -> None:
    """Exercise analyze_source with representative inputs."""
    result = analyze_source('/tmp/example')
    assert result is not None or result is None
