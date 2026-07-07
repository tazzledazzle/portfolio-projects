from src.enrich.cwe_mapper import extract_cwe
from src.format.comment_formatter import format_comment


def build_comment_payload(finding: dict, severity: str, playbook_url: str) -> dict:
    cwe = extract_cwe(finding.get("check_id", ""))
    body = format_comment(finding, severity, cwe, playbook_url)
    return {"body": body, "path": finding.get("path"), "line": finding.get("start", {}).get("line")}
