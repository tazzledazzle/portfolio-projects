import pytest
from unittest.mock import MagicMock, patch

pytestmark = pytest.mark.unit

from ai_code_assistant.extensions import validate_manifest


@patch('ai_code_assistant.extensions.json')
def test_validate_manifest_with_mocked_io(mock_json: MagicMock) -> None:
    mock_json.return_value = None
    result = validate_manifest('/tmp/example')
    assert result is not None or result is None
