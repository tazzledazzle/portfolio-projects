from friction_scorer.adapters.ci_source import sample_ci_metrics
from friction_scorer.core.composite import compute_composite_score


def main() -> None:
    metrics = sample_ci_metrics()
    score = compute_composite_score(metrics)
    print({"team": "platform", "friction_score": score})


if __name__ == "__main__":
    main()
