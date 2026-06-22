from pathlib import Path
import sys

sys.path.append(str(Path(__file__).resolve().parents[1] / "src"))

from main import render_pipeline


def test_render_pipeline_contains_core_stages() -> None:
    output = render_pipeline("svc", "prod")
    assert "lint:" in output
    assert "test:" in output
    assert "build:" in output
    assert "deploy:" in output
