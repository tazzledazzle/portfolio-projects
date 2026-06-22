from __future__ import annotations


def flakiness_score(outcomes: list[str]) -> float:
    if not outcomes:
        return 0.0
    failures = sum(1 for o in outcomes if o == "fail")
    return failures / len(outcomes)


def gate_decision(outcomes: list[str], threshold: float = 0.3) -> dict:
    score = flakiness_score(outcomes)
    return {
        "flakiness_score": round(score, 4),
        "decision": "allow" if score >= threshold else "block",
        "reason": "likely flaky" if score >= threshold else "likely regression",
    }


def main() -> None:
    sample = ["pass", "fail", "pass", "fail", "pass"]
    print(gate_decision(sample))


if __name__ == "__main__":
    main()
