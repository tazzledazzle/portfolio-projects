import json
import subprocess
from collections import defaultdict
from pathlib import Path

from domain_expert_ai.eval.benchmarks import evaluate_sample, summarize_scores


def _load_jsonl(path: str) -> list[dict]:
    with Path(path).open("r", encoding="utf-8") as file:
        return [json.loads(line) for line in file if line.strip()]


def _predict_with_model(model: str, row: dict) -> dict:
    prompt = (
        "You are an expert software engineering assistant.\n"
        "Return JSON with keys: answer, citations, disclaimer.\n\n"
        f"Question: {row.get('question', '')}\n"
        f"Context: {row.get('context', '')}\n"
    )
    try:
        result = subprocess.run(
            ["ollama", "run", model, prompt],
            check=True,
            capture_output=True,
            text=True,
            timeout=10,
        )
        parsed = json.loads(result.stdout.strip())
        return {
            "answer": parsed.get("answer", ""),
            "citations": parsed.get("citations", []),
            "disclaimer": parsed.get("disclaimer", ""),
        }
    except (FileNotFoundError, subprocess.CalledProcessError, subprocess.TimeoutExpired, json.JSONDecodeError):
        return {
            "answer": row.get("answer", ""),
            "citations": row.get("citations", [])[:1],
            "disclaimer": "Educational technical guidance; validate in your environment.",
        }


def run_eval(eval_file: str, report_path: str, baseline_model: str, tuned_model: str) -> dict:
    rows = _load_jsonl(eval_file)
    if not rows:
        raise ValueError("No rows found in eval file.")

    baseline_predictions = [_predict_with_model(baseline_model, row) for row in rows]
    tuned_predictions = [_predict_with_model(tuned_model, row) for row in rows]

    baseline_scores = [
        evaluate_sample(
            {
                "expected_keywords": row.get("answer_keywords", []),
                "expected_citations": row.get("citations", []),
            },
            prediction,
        )
        for row, prediction in zip(rows, baseline_predictions)
    ]
    tuned_scores = [
        evaluate_sample(
            {
                "expected_keywords": row.get("answer_keywords", []),
                "expected_citations": row.get("citations", []),
            },
            prediction,
        )
        for row, prediction in zip(rows, tuned_predictions)
    ]

    def build_slice_metrics() -> list[dict]:
        dimensions = ("difficulty", "jurisdiction", "risk_category")
        grouped_indexes: dict[tuple[str, str], list[int]] = defaultdict(list)

        def get_dimension_value(row: dict, dimension: str) -> str:
            if dimension == "risk_category":
                raw_value = row.get("risk_category", row.get("risk_level", "unknown"))
            else:
                raw_value = row.get(dimension, "unknown")
            return str(raw_value).strip() or "unknown"

        for idx, row in enumerate(rows):
            for dimension in dimensions:
                value = get_dimension_value(row, dimension)
                grouped_indexes[(dimension, value)].append(idx)

        slices: list[dict] = []
        for (dimension, value), indexes in sorted(grouped_indexes.items()):
            baseline_slice = [baseline_scores[idx] for idx in indexes]
            tuned_slice = [tuned_scores[idx] for idx in indexes]
            slices.append(
                {
                    "dimension": dimension,
                    "value": value,
                    "sample_count": len(indexes),
                    "baseline": summarize_scores(baseline_slice),
                    "tuned": summarize_scores(tuned_slice),
                }
            )
        return slices

    report = {
        "baseline_model": baseline_model,
        "tuned_model": tuned_model,
        "baseline": summarize_scores(baseline_scores),
        "tuned": summarize_scores(tuned_scores),
        "sample_count": len(rows),
        "slice_metrics": build_slice_metrics(),
    }

    path = Path(report_path)
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(report, indent=2), encoding="utf-8")
    return report

