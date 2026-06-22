import re
from difflib import SequenceMatcher


_TECH_REFERENCE_PATTERN = re.compile(
    r"^\s*(RFC\s*\d+|PEP\s*\d+|IEEE\s+\d+|ISO/[A-Z]+\s*\d+|"
    r"CLRS(\s+Chapter\s+\d+)?|Designing Data-Intensive Applications|"
    r"Deep Learning\s+\(Goodfellow et al\.\)|Kubernetes Docs|PyTorch Docs|TensorFlow Docs|"
    r"[A-Za-z][A-Za-z0-9 .:\-_/()]{4,}\s+(Docs|Documentation|Guide|Paper|Chapter)\b.*)\s*$",
    flags=re.IGNORECASE,
)


def citation_issues(citations: list[str]) -> list[str]:
    cleaned = [str(item).strip() for item in citations if str(item).strip()]
    issues: list[str] = []

    if not cleaned:
        issues.append("at least one citation is required")
        return issues

    valid_count = 0
    for citation in cleaned:
        if _TECH_REFERENCE_PATTERN.match(citation):
            valid_count += 1
        else:
            issues.append(f"invalid citation format: {citation}")
    if valid_count == 0:
        issues.insert(0, "at least one valid citation is required")
    return issues


def is_near_duplicate(first: str, second: str, threshold: float = 0.8) -> bool:
    if threshold < 0.0 or threshold > 1.0:
        raise ValueError("threshold must be between 0.0 and 1.0")
    left = " ".join(first.lower().split())
    right = " ".join(second.lower().split())
    return SequenceMatcher(None, left, right).ratio() >= threshold
