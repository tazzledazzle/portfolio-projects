from pathlib import Path

from ai_code_assistant.services.ast_analysis import analyze_source


def test_analyze_source_extracts_symbols(tmp_path: Path) -> None:
    src = tmp_path / "sample.py"
    src.write_text(
        "import os\nfrom math import sqrt\n\nclass C:\n    pass\n\nasync def af():\n    return 1\n\ndef f():\n    return 2\n",
        encoding="utf-8",
    )
    facts = analyze_source(src)

    assert facts.module_name == "sample"
    assert facts.function_names == ["af", "f"]
    assert facts.class_names == ["C"]
    assert "os" in facts.import_names or "sqrt" in facts.import_names
    assert facts.has_async is True
