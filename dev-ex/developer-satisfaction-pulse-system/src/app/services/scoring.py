def rolling_nps(scores: list[int]) -> float:
    if not scores:
        return 0.0

    promoters = sum(1 for score in scores if score >= 9)
    detractors = sum(1 for score in scores if score <= 6)
    total = len(scores)
    return ((promoters - detractors) / total) * 100.0
