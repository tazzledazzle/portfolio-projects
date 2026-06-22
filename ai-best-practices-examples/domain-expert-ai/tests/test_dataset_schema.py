from pydantic import ValidationError

from domain_expert_ai.data.prepare_dataset import prepare_dataset
from domain_expert_ai.data.schema import DatasetRecord


def test_dataset_record_accepts_valid_row():
    record = DatasetRecord(
        question="How does binary search achieve logarithmic time?",
        context="Sorted integer array with random access.",
        answer="Binary search halves the search interval each step, yielding O(log n).",
        jurisdiction="CS-DSA",
        risk_level="medium",
        citations=["CLRS Chapter 3"],
    )

    assert record.jurisdiction == "CS-DSA"
    assert record.risk_level == "medium"


def test_dataset_record_requires_citation():
    try:
        DatasetRecord(
            question="How should I tune gradient descent learning rate?",
            context="Training loss oscillates during optimization.",
            answer="Use smaller step sizes or schedule decay.",
            jurisdiction="ML-AI",
            risk_level="high",
            citations=[],
        )
    except ValidationError as exc:
        assert "citations" in str(exc)
    else:
        raise AssertionError("Expected validation error for empty citations.")


def test_prepare_dataset_filters_quality_and_writes_report(tmp_path):
    input_path = tmp_path / "input.jsonl"
    train_output = tmp_path / "train.jsonl"
    val_output = tmp_path / "val.jsonl"
    report_output = tmp_path / "quality_report.json"

    input_path.write_text(
        "\n".join(
            [
                    '{"question":"Can I optimize this array algorithm time complexity?","context":"Algorithm in arrays over large datasets.","answer":"Use stronger data structures and avoid nested scans.","jurisdiction":"CS-DSA","risk_level":"medium","citations":["CLRS Chapter 7"]}',
                    '{"question":"Can I optimize this array algorithm time complexity??","context":"Algorithm in arrays over large datasets.","answer":"Use stronger data structures and avoid nested scans.","jurisdiction":"CS-DSA","risk_level":"medium","citations":["CLRS Chapter 7"]}',
                '{"question":"How do I reduce p99 latency in a microservice?","context":"High tail latency under burst traffic.","answer":"Profile first and apply backpressure.","jurisdiction":"SE-PERF","risk_level":"low","citations":["Designing Data-Intensive Applications"]}',
                '{"question":"How do I tune distributed retries safely?","context":"Service retries can amplify outages.","answer":"Use bounded retries with jitter and circuit breakers.","jurisdiction":"SE-RELIABILITY","risk_level":"high","citations":["Company handbook"]}',
            ]
        )
        + "\n",
        encoding="utf-8",
    )

    train_count, val_count = prepare_dataset(
        str(input_path),
        str(train_output),
        str(val_output),
        val_ratio=0.5,
        seed=42,
        quality_report_path=str(report_output),
        enable_augmentation=False,
    )

    assert train_count == 1
    assert val_count == 1

    report = report_output.read_text(encoding="utf-8")
    assert '"accepted_records": 2' in report
    assert '"rejected_records": 2' in report
    assert '"duplicate_question": 1' in report
    assert '"citation_issue": 1' in report


def test_prepare_dataset_rejects_invalid_val_ratio(tmp_path):
    input_path = tmp_path / "input.jsonl"
    input_path.write_text(
        '{"question":"Q valid question one?","context":"Context ok.","answer":"Answer text enough.","jurisdiction":"CS-DSA","risk_level":"low","citations":["CLRS Chapter 1"]}\n'
        '{"question":"Q valid question two?","context":"Context ok.","answer":"Another answer text enough.","jurisdiction":"SE-PERF","risk_level":"low","citations":["RFC 9110"]}\n',
        encoding="utf-8",
    )

    try:
        prepare_dataset(
            str(input_path),
            str(tmp_path / "train.jsonl"),
            str(tmp_path / "val.jsonl"),
            val_ratio=1.0,
        )
    except ValueError as exc:
        assert "val_ratio" in str(exc)
    else:
        raise AssertionError("Expected ValueError for invalid val_ratio")


def test_prepare_dataset_counts_invalid_json_in_quality_report(tmp_path):
    input_path = tmp_path / "input.jsonl"
    report_output = tmp_path / "quality_report.json"
    input_path.write_text(
        '{"question":"Can I use an LRU cache for this workload?","context":"API read-heavy traffic.","answer":"Yes with proper eviction sizing.","jurisdiction":"CS-DSA","risk_level":"low","citations":["CLRS Chapter 10"]}\n'
        '{"bad_json":\n'
        '{"question":"How do I detect concept drift in production?","context":"Model quality degrades over time.","answer":"Track feature and label drift metrics.","jurisdiction":"ML-AI","risk_level":"low","citations":["Google MLOps documentation"]}\n',
        encoding="utf-8",
    )

    train_count, val_count = prepare_dataset(
        str(input_path),
        str(tmp_path / "train.jsonl"),
        str(tmp_path / "val.jsonl"),
        val_ratio=0.5,
        quality_report_path=str(report_output),
    )
    assert train_count + val_count == 2
    report = report_output.read_text(encoding="utf-8")
    assert '"invalid_json": 1' in report

