from domain_expert_ai.data.curation.augment_cases import augment_case_variants
from domain_expert_ai.data.curation.quality_checks import (
    citation_issues,
    is_near_duplicate,
)


def test_citation_issues_flags_missing_and_bad_format():
    issues = citation_issues(["", "invalid citation"])
    assert "at least one valid citation is required" in issues
    assert "invalid citation format: invalid citation" in issues


def test_citation_issues_flags_missing_citations():
    issues = citation_issues(["", " "])
    assert issues == ["at least one citation is required"]


def test_citation_issues_accepts_reasonable_technical_reference():
    issues = citation_issues(["CLRS Chapter 7"])
    assert issues == []


def test_is_near_duplicate_detects_small_wording_changes():
    first = "When should I use quicksort for large arrays?"
    second = "When should I use quicksort for big arrays?"
    assert is_near_duplicate(first, second) is True


def test_is_near_duplicate_rejects_different_topics():
    first = "How does Dijkstra algorithm work?"
    second = "How do I reduce overfitting in neural networks?"
    assert is_near_duplicate(first, second) is False


def test_is_near_duplicate_rejects_invalid_threshold():
    try:
        is_near_duplicate("a", "b", threshold=1.5)
    except ValueError as exc:
        assert "threshold" in str(exc)
    else:
        raise AssertionError("Expected ValueError for invalid threshold")


def test_augment_case_variants_creates_state_and_fact_variants():
    record = {
        "question": "How do I optimize time complexity for this array algorithm?",
        "context": "Current array approach is O(n^2) on large inputs.",
        "answer": "Use better data structures and reduce nested scans.",
        "jurisdiction": "CS-DSA",
        "citations": ["CLRS Chapter 7"],
    }

    variants = augment_case_variants(record)
    assert len(variants) == 2

    concept_variant = next(item for item in variants if item["variant_type"] == "concept")
    assert "space complexity" in concept_variant["question"].lower()
    assert "linked list" in concept_variant["context"].lower()

    fact_variant = next(item for item in variants if item["variant_type"] == "fact")
    assert "low latency" in fact_variant["context"].lower()


def test_augment_case_variants_does_not_mutate_original_record():
    record = {
        "question": "How do I optimize time complexity for this array algorithm?",
        "context": "Current array approach is O(n^2) on large inputs.",
        "answer": "Use better data structures and reduce nested scans.",
        "jurisdiction": "CS-DSA",
        "citations": ["CLRS Chapter 7"],
    }
    original_context = record["context"]

    _ = augment_case_variants(record)
    assert record["context"] == original_context


def test_augment_case_variants_non_ca_only_has_fact_variant():
    record = {
        "question": "How do I harden a distributed cache?",
        "context": "Service has cache stampedes during traffic spikes.",
        "answer": "Use request coalescing and jittered TTL.",
        "jurisdiction": "SE-PERF",
        "citations": ["Designing Data-Intensive Applications"],
    }

    variants = augment_case_variants(record)
    assert len(variants) == 1
    assert variants[0]["variant_type"] == "fact"
