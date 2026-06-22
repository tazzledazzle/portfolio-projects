from src.format.comment_formatter import format_comment


def test_format_comment_contains_cwe() -> None:
    result = format_comment({"extra": {"message": "Potential SQL injection"}}, "high", "cwe-89", "https://playbook")
    assert "cwe-89" in result
