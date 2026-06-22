from __future__ import annotations

from statistics import mean


def calculate_lead_times_hours(events: list[dict]) -> list[float]:
    return [float(e["deploy_hours_after_merge"]) for e in events]


def summarize(times: list[float]) -> dict:
    if not times:
        return {"count": 0, "mean": 0.0, "p50": 0.0, "p90": 0.0}
    ordered = sorted(times)
    p50_idx = int(0.5 * (len(ordered) - 1))
    p90_idx = int(0.9 * (len(ordered) - 1))
    return {
        "count": len(times),
        "mean": round(mean(times), 2),
        "p50": ordered[p50_idx],
        "p90": ordered[p90_idx],
    }


def main() -> None:
    sample = [
        {"deploy_hours_after_merge": 2},
        {"deploy_hours_after_merge": 6},
        {"deploy_hours_after_merge": 10},
    ]
    print(summarize(calculate_lead_times_hours(sample)))


if __name__ == "__main__":
    main()
