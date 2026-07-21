from authz import authorize


def test_authorize_accepts_read_scope_token():
    result = authorize("Bearer demo-reader:tools:read", {"tools:read"})

    assert result.allowed is True
    assert result.subject == "demo-reader"


def test_authorize_denies_missing_or_invalid_token():
    for token in (None, "", "Bearer invalid", "Basic demo-reader:tools:read"):
        result = authorize(token, {"tools:read"})
        assert result.allowed is False


def test_authorize_denies_missing_scope():
    result = authorize("Bearer demo-reader:profile:read", {"tools:read"})

    assert result.allowed is False
    assert result.reason == "missing_scope"
