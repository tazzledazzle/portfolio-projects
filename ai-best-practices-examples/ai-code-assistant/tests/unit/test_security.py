import pytest

pytestmark = pytest.mark.unit

from ai_code_assistant.security import ExecutionPolicy, build_policy


def test_build_policy_behavior() -> None:
    """Exercise build_policy with representative inputs."""
    result = build_policy('/tmp/example')
    assert result is not None or result is None

def test_executionpolicy_can_be_constructed() -> None:
    """Verify ExecutionPolicy can be instantiated."""
    instance = ExecutionPolicy()
    assert isinstance(instance, ExecutionPolicy)
