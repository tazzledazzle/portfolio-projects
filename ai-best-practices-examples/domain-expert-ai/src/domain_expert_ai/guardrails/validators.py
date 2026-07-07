from pydantic import BaseModel, Field


class GuardrailResult(BaseModel):
    ok: bool
    errors: list[str] = Field(default_factory=list)
    sanitized: dict = Field(default_factory=dict)


def validate_response_payload(payload: dict) -> GuardrailResult:
    errors: list[str] = []

    answer = str(payload.get("answer", "")).strip()
    if len(answer) < 16:
        errors.append("answer must be at least 16 characters")

    citations = payload.get("citations", [])
    if not isinstance(citations, list) or not any(str(item).strip() for item in citations):
        errors.append("citations must include at least one citation")

    confidence_value = payload.get("confidence", 0.0)
    try:
        confidence = float(confidence_value)
    except (TypeError, ValueError):
        errors.append("confidence must be numeric")
        confidence = 0.0
    else:
        if confidence < 0.0 or confidence > 1.0:
            errors.append("confidence must be between 0 and 1")

    disclaimer = str(payload.get("disclaimer", "")).strip()
    if not disclaimer:
        errors.append("disclaimer is required")

    sanitized = {
        "answer": answer,
        "citations": [str(item).strip() for item in citations if str(item).strip()],
        "confidence": confidence,
        "disclaimer": disclaimer,
    }
    return GuardrailResult(ok=not errors, errors=errors, sanitized=sanitized)

