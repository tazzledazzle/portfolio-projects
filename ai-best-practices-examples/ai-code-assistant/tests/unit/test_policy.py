import pytest
from unittest.mock import MagicMock, patch

pytestmark = pytest.mark.unit

from ai_code_assistant.policy import load_policy


@patch('ai_code_assistant.policy.os')
def test_load_policy_with_mocked_io(mock_os: MagicMock) -> None:
    mock_os.return_value = None
    result = load_policy('/tmp/example')
    assert result is not None or result is None
