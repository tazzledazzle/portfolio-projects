from __future__ import annotations


def evaluate_slo(error_rate: float, p99_latency_ms: int) -> bool:
    return error_rate <= 0.02 and p99_latency_ms <= 500


def rollout_step(weight_percent: int, error_rate: float, p99_latency_ms: int) -> dict:
    healthy = evaluate_slo(error_rate, p99_latency_ms)
    if not healthy:
        return {"action": "rollback", "next_weight": 0}
    next_weight = min(weight_percent + 20, 100)
    action = "promote" if next_weight == 100 else "advance"
    return {"action": action, "next_weight": next_weight}


def main() -> None:
    print(rollout_step(20, error_rate=0.01, p99_latency_ms=320))


if __name__ == "__main__":
    main()
