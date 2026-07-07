def extract_cwe(check_id: str) -> str:
    # Example: "python.flask.security.xss.cwe-79"
    parts = check_id.lower().split(".")
    for part in parts:
        if part.startswith("cwe-"):
            return part
    return "cwe-unknown"
