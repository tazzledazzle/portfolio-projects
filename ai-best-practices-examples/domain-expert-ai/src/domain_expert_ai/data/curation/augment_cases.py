from copy import deepcopy
import re


_TOPIC_SWAPS = [
    (r"\btime complexity\b", "space complexity"),
    (r"\barray\b", "linked list"),
    (r"\bcnn\b", "transformer"),
]


def augment_case_variants(record: dict) -> list[dict]:
    variants: list[dict] = []

    concept_variant = deepcopy(record)
    concept_variant["variant_type"] = "concept"
    question = str(concept_variant.get("question", ""))
    context = str(concept_variant.get("context", ""))
    swaps = 0
    for pattern, replacement in _TOPIC_SWAPS:
        question, q_count = re.subn(pattern, replacement, question, flags=re.IGNORECASE)
        context, c_count = re.subn(pattern, replacement, context, flags=re.IGNORECASE)
        swaps += q_count + c_count
    if swaps > 0:
        concept_variant["question"] = question
        concept_variant["context"] = context
        variants.append(concept_variant)

    fact_variant = deepcopy(record)
    fact_variant["variant_type"] = "fact"
    fact_variant["context"] = (
        f"{str(fact_variant.get('context', '')).strip()} "
        "Assume constraints include high scale, low latency, and limited memory."
    ).strip()
    variants.append(fact_variant)

    return variants
