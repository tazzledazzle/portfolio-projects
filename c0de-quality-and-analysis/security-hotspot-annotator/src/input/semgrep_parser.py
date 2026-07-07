def parse_semgrep(findings: dict) -> list[dict]:
    return findings.get("results", [])
