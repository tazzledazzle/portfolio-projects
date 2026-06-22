from ai_code_assistant.security import build_policy


def test_read_only_policy_cannot_write() -> None:
    assert build_policy("read-only").can_write is False


def test_workspace_write_policy_can_write() -> None:
    assert build_policy("workspace-write").can_write is True


def test_full_access_policy_can_write() -> None:
    assert build_policy("full-access").can_write is True
