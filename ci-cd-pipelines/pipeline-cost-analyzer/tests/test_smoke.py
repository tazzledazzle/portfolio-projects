from pathlib import Path
import sys

sys.path.append(str(Path(__file__).resolve().parents[1] / "src"))

from main import attributed_cost, recommend


def test_attributed_cost_rolls_up_by_team() -> None:
    totals = attributed_cost(
        [
            {"team": "core", "runner": "linux", "minutes": 10, "job": "lint"},
            {"team": "core", "runner": "windows", "minutes": 10, "job": "test"},
        ]
    )
    assert totals["core"] > 0


def test_recommend_returns_guidance_for_high_minutes() -> None:
    items = recommend([{"team": "core", "runner": "macos", "minutes": 31, "job": "ui-tests"}])
    assert len(items) >= 1
