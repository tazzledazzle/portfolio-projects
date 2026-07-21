from approval import ApprovalStore, intent_digest


def test_intent_digest_stable():
    first = intent_digest("restart_pipeline", {"pipeline": "build-42"})
    second = intent_digest("restart_pipeline", {"pipeline": "build-42"})
    changed = intent_digest("restart_pipeline", {"pipeline": "build-43"})

    assert first == second
    assert first != changed


def test_grant_validates_token_bound_to_digest():
    store = ApprovalStore()
    approved = intent_digest("restart_pipeline", {"pipeline": "build-42"})
    different = intent_digest("restart_pipeline", {"pipeline": "build-43"})

    token = store.grant(approved)

    assert token.startswith("appr_")
    assert store.valid(token, approved) is True
    assert store.valid(token, different) is False
    assert store.valid(None, approved) is False
