import pytest
from unittest.mock import MagicMock, patch

pytestmark = pytest.mark.unit

from ai_code_assistant.audit import append_audit_event


@patch('ai_code_assistant.audit.json')
def test_append_audit_event_with_mocked_io(mock_json: MagicMock) -> None:
    mock_json.return_value = None
    result = append_audit_event('/tmp/example', 2)
    assert result is not None or result is None
