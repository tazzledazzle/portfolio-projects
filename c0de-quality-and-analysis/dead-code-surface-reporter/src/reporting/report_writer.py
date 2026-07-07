def to_markdown(backlog_items: list[dict]) -> str:
    lines = ["# Dead Code Backlog", ""]
    for item in backlog_items:
        lines.append(f"- `{item['symbol']}` score={item['score']}")
    return "\n".join(lines)
