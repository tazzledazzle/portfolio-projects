from domain_expert_ai.eval.benchmarks import evaluate_sample
from domain_expert_ai.eval.run_eval import run_eval


def test_evaluate_sample_scores_format_and_citations():
    sample = {
        "question": "When should I use binary search?",
        "expected_keywords": ["sorted", "log n"],
        "expected_citations": ["CLRS Chapter 3"],
    }
    prediction = {
        "answer": "Use binary search on sorted arrays to get O(log n) lookup.",
        "citations": ["CLRS Chapter 3"],
        "disclaimer": "Educational technical guidance; validate in your environment.",
    }
    result = evaluate_sample(sample, prediction)
    assert result["keyword_score"] == 1.0
    assert result["citation_score"] == 1.0
    assert result["format_ok"] is True


def test_evaluate_sample_flags_missing_disclaimer():
    sample = {
        "question": "How do I reduce model overfitting?",
        "expected_keywords": ["regularization"],
        "expected_citations": ["Deep Learning (Goodfellow et al.)"],
    }
    prediction = {"answer": "Try regularization and early stopping.", "citations": []}
    result = evaluate_sample(sample, prediction)
    assert result["format_ok"] is False


def test_run_eval_handles_rows_without_answer_keywords(tmp_path):
    eval_file = tmp_path / "eval.jsonl"
    report_file = tmp_path / "report.json"
    eval_file.write_text(
        '{"answer":"Use hash maps for average O(1) key lookup.","citations":["CLRS Chapter 11"]}\n',
        encoding="utf-8",
    )

    report = run_eval(str(eval_file), str(report_file), "baseline-a", "tuned-b")
    assert report["sample_count"] == 1
    assert report_file.exists()


def test_run_eval_includes_slice_metrics_and_keeps_global_shape(tmp_path, monkeypatch):
    eval_file = tmp_path / "eval.jsonl"
    report_file = tmp_path / "report.json"
    eval_file.write_text(
        "\n".join(
            [
                '{"question":"Q1","context":"C1","answer_keywords":["duties test"],'
                '"citations":["CLRS Chapter 3"],"difficulty":"easy",'
                '"jurisdiction":"CS-DSA","risk_level":"medium"}',
                '{"question":"Q2","context":"C2","answer_keywords":["gradient clipping"],'
                '"citations":["PyTorch Docs"],"difficulty":"hard",'
                '"jurisdiction":"ML-AI","risk_category":"model-stability"}',
            ]
        )
        + "\n",
        encoding="utf-8",
    )

    def fake_predict(model: str, row: dict) -> dict:
        if model == "baseline-a":
            return {
                "answer": "generic answer",
                "citations": [],
                "disclaimer": "Educational technical guidance; validate in your environment.",
            }
        return {
            "answer": "Includes sorted arrays and gradient clipping guidance",
            "citations": ["CLRS Chapter 3", "PyTorch Docs"],
            "disclaimer": "Educational technical guidance; validate in your environment.",
        }

    monkeypatch.setattr("domain_expert_ai.eval.run_eval._predict_with_model", fake_predict)

    report = run_eval(str(eval_file), str(report_file), "baseline-a", "tuned-b")

    # Backward-compatible global metrics still exist.
    assert "baseline" in report
    assert "tuned" in report
    assert "sample_count" in report
    assert report["sample_count"] == 2

    # New slice metrics table exists for required slice dimensions.
    assert "slice_metrics" in report
    assert isinstance(report["slice_metrics"], list)
    dimensions = {row["dimension"] for row in report["slice_metrics"]}
    assert dimensions == {"difficulty", "jurisdiction", "risk_category"}
    risk_values = {row["value"] for row in report["slice_metrics"] if row["dimension"] == "risk_category"}
    assert "medium" in risk_values

