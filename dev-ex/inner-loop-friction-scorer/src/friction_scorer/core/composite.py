def compute_composite_score(metrics: dict[str, float]) -> float:
    weighted = (
        metrics.get("setup_time", 0.0) * 0.25
        + metrics.get("build_time", 0.0) * 0.2
        + metrics.get("test_time", 0.0) * 0.2
        + metrics.get("review_latency", 0.0) * 0.2
        + metrics.get("deploy_time", 0.0) * 0.15
    )
    return round(weighted, 2)
