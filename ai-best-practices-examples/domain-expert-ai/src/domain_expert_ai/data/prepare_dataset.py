import json
import random
from pathlib import Path

from pydantic import ValidationError

from domain_expert_ai.data.curation.augment_cases import augment_case_variants
from domain_expert_ai.data.curation.quality_checks import citation_issues, is_near_duplicate
from domain_expert_ai.data.schema import DatasetRecord


def _load_raw_rows(input_path: str) -> tuple[list[dict], int]:
    rows: list[dict] = []
    invalid_json = 0
    with Path(input_path).open("r", encoding="utf-8") as file:
        for line in file:
            if not line.strip():
                continue
            try:
                rows.append(json.loads(line))
            except json.JSONDecodeError:
                invalid_json += 1
    return rows, invalid_json


def _dump_records(output_path: str, records: list[DatasetRecord]) -> None:
    path = Path(output_path)
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("w", encoding="utf-8") as file:
        for record in records:
            file.write(json.dumps(record.model_dump(), ensure_ascii=True) + "\n")


def _dump_quality_report(output_path: str, report: dict) -> None:
    path = Path(output_path)
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("w", encoding="utf-8") as file:
        json.dump(report, file, ensure_ascii=True, indent=2)


def _build_curated_records(rows: list[dict], invalid_json_count: int = 0) -> tuple[list[DatasetRecord], dict]:
    accepted: list[DatasetRecord] = []
    rejection_reasons: dict[str, int] = {}
    seen_questions: list[str] = []

    for row in rows:
        reasons: list[str] = []
        try:
            record = DatasetRecord.model_validate(row)
        except ValidationError:
            reasons.append("invalid_schema")
            record = None

        if record is not None:
            citation_problems = citation_issues(record.citations)
            if citation_problems:
                reasons.append("citation_issue")

            duplicate_found = any(is_near_duplicate(record.question, seen) for seen in seen_questions)
            if duplicate_found:
                reasons.append("duplicate_question")

        if reasons:
            for reason in reasons:
                rejection_reasons[reason] = rejection_reasons.get(reason, 0) + 1
            continue

        accepted.append(record)
        seen_questions.append(record.question)

    report = {
        "input_rows": len(rows) + invalid_json_count,
        "accepted_records": len(accepted),
        "rejected_records": (len(rows) - len(accepted)) + invalid_json_count,
        "rejection_reasons": rejection_reasons,
    }
    if invalid_json_count:
        report["rejection_reasons"]["invalid_json"] = invalid_json_count
    return accepted, report


def prepare_dataset(
    input_path: str,
    train_output: str,
    val_output: str,
    val_ratio: float = 0.2,
    seed: int = 42,
    quality_report_path: str | None = None,
    enable_augmentation: bool = False,
) -> tuple[int, int]:
    if val_ratio <= 0 or val_ratio >= 1:
        raise ValueError("val_ratio must be between 0 and 1 (exclusive).")

    rows, invalid_json_count = _load_raw_rows(input_path)
    if enable_augmentation:
        augmented_rows: list[dict] = []
        for row in rows:
            augmented_rows.extend(augment_case_variants(row))
        rows = rows + augmented_rows

    records, report = _build_curated_records(rows, invalid_json_count=invalid_json_count)
    if quality_report_path:
        _dump_quality_report(quality_report_path, report)

    if len(records) < 2:
        raise ValueError("Need at least 2 valid rows to build train/val split.")

    random.Random(seed).shuffle(records)
    val_size = max(1, int(len(records) * val_ratio))
    val_records = records[:val_size]
    train_records = records[val_size:]
    if not train_records:
        raise ValueError("Validation ratio too large; train split is empty.")

    _dump_records(train_output, train_records)
    _dump_records(val_output, val_records)
    return len(train_records), len(val_records)

