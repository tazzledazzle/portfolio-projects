import ast
from dataclasses import dataclass
from pathlib import Path


@dataclass(frozen=True)
class AnalysisFacts:
    module_name: str
    function_names: list[str]
    class_names: list[str]
    import_names: list[str]
    has_async: bool


def analyze_source(source_path: Path) -> AnalysisFacts:
    source_code = source_path.read_text(encoding="utf-8")
    tree = ast.parse(source_code)
    function_names: list[str] = []
    class_names: list[str] = []
    import_names: list[str] = []
    has_async = False

    for node in ast.walk(tree):
        if isinstance(node, ast.FunctionDef):
            function_names.append(node.name)
        elif isinstance(node, ast.AsyncFunctionDef):
            function_names.append(node.name)
            has_async = True
        elif isinstance(node, ast.ClassDef):
            class_names.append(node.name)
        elif isinstance(node, ast.Import):
            import_names.extend(alias.name for alias in node.names)
        elif isinstance(node, ast.ImportFrom):
            import_names.extend(alias.name for alias in node.names)

    return AnalysisFacts(
        module_name=source_path.stem,
        function_names=sorted(set(function_names)),
        class_names=sorted(set(class_names)),
        import_names=sorted(set(import_names)),
        has_async=has_async,
    )
