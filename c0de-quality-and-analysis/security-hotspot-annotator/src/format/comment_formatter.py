def format_comment(finding: dict, severity: str, cwe: str, playbook_url: str) -> str:
    return (
        f"**{severity.upper()}** security hotspot ({cwe}). "
        f"Risk: {finding.get('extra', {}).get('message', 'No message provided')}. "
        f"Remediation: {playbook_url}"
    )
