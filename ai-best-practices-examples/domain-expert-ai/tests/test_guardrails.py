from domain_expert_ai.guardrails.validators import validate_response_payload


def test_validate_response_payload_rejects_missing_disclaimer():
    payload = {
        "answer": "Use a bounded queue and backpressure to protect tail latency.",
        "citations": ["Designing Data-Intensive Applications"],
        "confidence": 0.79,
        "disclaimer": "",
    }
    result = validate_response_payload(payload)
    assert result.ok is False
    assert "disclaimer" in result.errors[0]


def test_validate_response_payload_accepts_valid_payload():
    payload = {
        "answer": "Quicksort is often fast in practice, but mergesort is stable.",
        "citations": ["CLRS Chapter 7"],
        "confidence": 0.84,
        "disclaimer": "Educational technical guidance; validate in your environment.",
    }
    result = validate_response_payload(payload)
    assert result.ok is True
    assert result.sanitized["confidence"] == 0.84

