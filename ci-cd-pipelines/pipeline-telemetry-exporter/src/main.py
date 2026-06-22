from __future__ import annotations

import json
from datetime import datetime, timezone


def _now_iso() -> str:
    return datetime.now(timezone.utc).isoformat()


def build_spans(event: dict) -> list[dict]:
    job = event.get("job", "unknown-job")
    steps = event.get("steps", [])
    spans = []
    for step in steps:
        spans.append(
            {
                "trace_id": event.get("trace_id", "bootstrap-trace"),
                "span_name": f"{job}:{step.get('name', 'unknown-step')}",
                "start_time": step.get("started_at", _now_iso()),
                "end_time": step.get("ended_at", _now_iso()),
                "attributes": {
                    "job": job,
                    "status": step.get("status", "unknown"),
                    "failure_reason": step.get("failure_reason", ""),
                },
            }
        )
    return spans


def main() -> None:
    sample_event = {
        "trace_id": "sample-trace",
        "job": "test-suite",
        "steps": [{"name": "unit-tests", "status": "passed"}],
    }
    print(json.dumps(build_spans(sample_event), indent=2))


if __name__ == "__main__":
    main()
