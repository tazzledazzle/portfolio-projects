from __future__ import annotations


RUNNER_COST_PER_MINUTE = {"linux": 0.008, "windows": 0.016, "macos": 0.08}


def attributed_cost(records: list[dict]) -> dict[str, float]:
    totals: dict[str, float] = {}
    for rec in records:
        team = rec["team"]
        rate = RUNNER_COST_PER_MINUTE[rec["runner"]]
        cost = rec["minutes"] * rate
        totals[team] = totals.get(team, 0.0) + cost
    return {k: round(v, 2) for k, v in totals.items()}


def recommend(records: list[dict]) -> list[str]:
    suggestions = []
    for rec in records:
        if rec["runner"] == "macos" and rec["minutes"] > 20:
            suggestions.append(f"Consider moving {rec['job']} off macOS runner.")
        if rec["minutes"] > 30:
            suggestions.append(f"Add caching/parallelism for {rec['job']}.")
    return suggestions


def main() -> None:
    sample = [{"team": "platform", "runner": "linux", "minutes": 120, "job": "integration"}]
    print(attributed_cost(sample))
    print(recommend(sample))


if __name__ == "__main__":
    main()
