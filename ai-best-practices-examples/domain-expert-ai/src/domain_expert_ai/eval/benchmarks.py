from typing import Any


def _normalize(text: str) -> str:
    return " ".join(text.lower().split())


def evaluate_sample(sample: dict[str, Any], prediction: dict[str, Any]) -> dict[str, Any]:
    answer = _normalize(str(prediction.get("answer", "")))
    keywords = [_normalize(value) for value in sample.get("expected_keywords", [])]
    expected_citations = [_normalize(value) for value in sample.get("expected_citations", [])]
    actual_citations = [_normalize(str(value)) for value in prediction.get("citations", [])]

    keyword_hits = sum(1 for kw in keywords if kw in answer)
    citation_hits = sum(1 for cite in expected_citations if cite in actual_citations)

    keyword_score = keyword_hits / max(1, len(keywords))
    citation_score = citation_hits / max(1, len(expected_citations))
    format_ok = bool(prediction.get("disclaimer", "") and isinstance(prediction.get("citations", []), list))

    return {
        "keyword_score": round(keyword_score, 4),
        "citation_score": round(citation_score, 4),
        "format_ok": format_ok,
    }


def summarize_scores(items: list[dict[str, Any]]) -> dict[str, float]:
    if not items:
        return {"keyword_score": 0.0, "citation_score": 0.0, "format_rate": 0.0}

    count = len(items)
    keyword_score = sum(float(item["keyword_score"]) for item in items) / count
    citation_score = sum(float(item["citation_score"]) for item in items) / count
    format_rate = sum(1.0 if item["format_ok"] else 0.0 for item in items) / count
    return {
        "keyword_score": round(keyword_score, 4),
        "citation_score": round(citation_score, 4),
        "format_rate": round(format_rate, 4),
    }

