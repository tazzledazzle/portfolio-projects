from pathlib import Path
import sys

sys.path.append(str(Path(__file__).resolve().parents[1] / "src"))

from main import calculate_lead_times_hours, summarize


def test_summarize_returns_expected_shape() -> None:
    times = calculate_lead_times_hours(
        [{"deploy_hours_after_merge": 1}, {"deploy_hours_after_merge": 5}]
    )
    result = summarize(times)
    assert result["count"] == 2
    assert "p90" in result
