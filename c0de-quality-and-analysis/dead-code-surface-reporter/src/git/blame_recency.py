from datetime import datetime, UTC


def stale_days(last_touched_iso: str) -> int:
    last_touch = datetime.fromisoformat(last_touched_iso)
    return (datetime.now(UTC) - last_touch).days
