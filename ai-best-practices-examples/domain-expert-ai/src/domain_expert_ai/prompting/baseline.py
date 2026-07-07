from typing import Any


BASELINE_SYSTEM_PROMPT = (
    "You are an expert software engineering and computer science assistant. "
    "Provide concise technical guidance, include at least one credible reference, "
    "and include a short disclaimer to validate choices in the target environment."
)


def build_baseline_prompt(question: str, context: str, few_shots: list[dict[str, str]]) -> str:
    shot_text = []
    for shot in few_shots:
        shot_text.append(f"Q: {shot['question']}\nA: {shot['answer']}")
    rendered_shots = "\n\n".join(shot_text)
    return (
        f"{BASELINE_SYSTEM_PROMPT}\n\n"
        f"Context:\n{context}\n\n"
        f"Examples:\n{rendered_shots}\n\n"
        f"Question:\n{question}\n\n"
        "Respond in JSON with keys: answer, citations, confidence, disclaimer."
    )


def parse_structured_response(raw: dict[str, Any]) -> dict[str, Any]:
    return {
        "answer": str(raw.get("answer", "")).strip(),
        "citations": [str(item).strip() for item in raw.get("citations", []) if str(item).strip()],
        "confidence": float(raw.get("confidence", 0.0)),
        "disclaimer": str(raw.get("disclaimer", "")).strip(),
    }

