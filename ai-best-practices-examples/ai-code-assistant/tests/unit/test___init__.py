import pytest

pytestmark = pytest.mark.unit

from ai_code_assistant.__init__ import *


def test___init___module_is_importable_unit() -> None:
    """Verify ai_code_assistant.__init__ imports without error."""
    import ai_code_assistant.__init__ as module
    assert module is not None
