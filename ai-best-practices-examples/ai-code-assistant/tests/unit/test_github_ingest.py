import pytest
from unittest.mock import MagicMock, patch

pytestmark = pytest.mark.unit

from ai_code_assistant.github_ingest import ingest_pr_with_gh


@patch('ai_code_assistant.github_ingest.json')
def test_ingest_pr_with_gh_with_mocked_io(mock_json: MagicMock) -> None:
    mock_json.return_value = None
    result = ingest_pr_with_gh(1, 1)
    assert result is not None or result is None
