def priority_score(unreachable_confidence: float, stale_days_count: int) -> float:
    normalized_days = min(stale_days_count / 365.0, 1.0)
    return round((0.6 * unreachable_confidence) + (0.4 * normalized_days), 3)
