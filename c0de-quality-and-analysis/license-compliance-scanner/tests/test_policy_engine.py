from src.policy.policy_engine import evaluate_policy


def test_restricted_license_is_flagged() -> None:
    violations = evaluate_policy(
        packages=[{"name": "pkg", "license": "GPL-3.0-only"}],
        allowed={"MIT", "Apache-2.0"},
        restricted={"GPL-3.0-only"},
    )
    assert len(violations) == 1
