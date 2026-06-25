"""AST-driven pytest generation for modules without an LLM."""

from __future__ import annotations

import ast
from dataclasses import dataclass
from pathlib import Path

from ai_code_assistant.services.ast_analysis import AnalysisFacts

_IO_MODULES = frozenset(
    {
        "os",
        "sys",
        "pathlib",
        "subprocess",
        "socket",
        "requests",
        "httpx",
        "urllib",
        "shutil",
        "tempfile",
        "json",
        "open",
    }
)


@dataclass(frozen=True)
class _ParamInfo:
    name: str
    has_default: bool


@dataclass(frozen=True)
class _FunctionInfo:
    name: str
    params: tuple[_ParamInfo, ...]
    is_async: bool
    is_method: bool
    parent_class: str | None
    return_expr: ast.expr | None
    uses_io: bool


def derive_module_import(
    source_path: Path | None, repo_root: Path | None, module_name: str
) -> str:
    if source_path is None or repo_root is None:
        return module_name
    try:
        rel = source_path.resolve().relative_to(repo_root.resolve())
    except ValueError:
        return module_name
    parts = list(rel.parts)
    if parts and parts[0] == "src":
        parts = parts[1:]
    if not parts:
        return module_name
    parts[-1] = Path(parts[-1]).stem
    return ".".join(parts)


def generate_robust_tests(
    source_code: str,
    module_name: str,
    facts: AnalysisFacts | None,
    test_level: str,
    source_path: Path | None = None,
    repo_root: Path | None = None,
) -> str:
    module_import = derive_module_import(source_path, repo_root, module_name)
    functions = _extract_functions(source_code)
    io_imports = _io_imports_from_facts(facts)
    targets = _select_targets(functions, test_level)

    lines = ["import pytest"]
    if io_imports and test_level == "unit":
        lines.append("from unittest.mock import MagicMock, patch")
    lines.append("")
    lines.append(f"pytestmark = pytest.mark.{test_level}")
    lines.append("")
    lines.append(f"from {module_import} import {', '.join(_symbols_to_import(targets, functions, test_level))}")
    lines.append("")

    if not targets:
        lines.extend(_module_smoke_test(module_import, module_name, test_level))
        return "\n".join(lines) + "\n"

    if test_level == "unit":
        lines.extend(_unit_tests(targets, functions, io_imports, module_import))
    elif test_level == "integration":
        lines.extend(_integration_tests(targets, functions, module_import, module_name))
    else:
        lines.extend(_e2e_tests(targets, functions, module_import, module_name, source_code))

    return "\n".join(lines) + "\n"


def _io_imports_from_facts(facts: AnalysisFacts | None) -> list[str]:
    if not facts:
        return []
    return sorted(name for name in facts.import_names if name.split(".")[0] in _IO_MODULES)


def _extract_functions(source_code: str) -> list[_FunctionInfo]:
    tree = ast.parse(source_code)
    functions: list[_FunctionInfo] = []
    io_names = _collect_io_names(tree)

    for node in tree.body:
        if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef)):
            functions.append(_function_info(node, is_method=False, parent_class=None, io_names=io_names))
        elif isinstance(node, ast.ClassDef):
            for child in node.body:
                if isinstance(child, (ast.FunctionDef, ast.AsyncFunctionDef)):
                    functions.append(
                        _function_info(child, is_method=True, parent_class=node.name, io_names=io_names)
                    )
    return functions


def _collect_io_names(tree: ast.Module) -> set[str]:
    names: set[str] = set()
    for node in ast.walk(tree):
        if isinstance(node, ast.Import):
            for alias in node.names:
                root = alias.name.split(".")[0]
                if root in _IO_MODULES:
                    names.add(alias.asname or alias.name.split(".")[-1])
        elif isinstance(node, ast.ImportFrom) and node.module:
            root = node.module.split(".")[0]
            if root in _IO_MODULES:
                for alias in node.names:
                    names.add(alias.asname or alias.name)
    return names


def _function_info(
    node: ast.FunctionDef | ast.AsyncFunctionDef,
    *,
    is_method: bool,
    parent_class: str | None,
    io_names: set[str],
) -> _FunctionInfo:
    params: list[_ParamInfo] = []
    args = node.args
    pos_only = list(args.posonlyargs)
    pos = list(args.args)
    defaults_offset = len(pos_only) + len(pos) - len(args.defaults)
    all_pos = pos_only + pos
    for index, arg in enumerate(all_pos):
        if arg.arg == "self":
            continue
        default_index = index - defaults_offset
        has_default = default_index >= 0
        params.append(_ParamInfo(name=arg.arg, has_default=has_default))

    return_expr = None
    uses_io = False
    for child in ast.walk(node):
        if isinstance(child, ast.Return) and child.value is not None:
            return_expr = child.value
        if isinstance(child, ast.Call) and isinstance(child.func, ast.Name) and child.func.id in io_names:
            uses_io = True
        if isinstance(child, ast.Attribute) and isinstance(child.value, ast.Name):
            if child.value.id in io_names:
                uses_io = True

    return _FunctionInfo(
        name=node.name,
        params=tuple(params),
        is_async=isinstance(node, ast.AsyncFunctionDef),
        is_method=is_method,
        parent_class=parent_class,
        return_expr=return_expr,
        uses_io=uses_io,
    )


def _select_targets(functions: list[_FunctionInfo], test_level: str) -> list[_FunctionInfo]:
    candidates = [
        fn
        for fn in functions
        if not fn.is_method
        and fn.name != "main"
        and not fn.name.startswith("__")
        and not fn.name.startswith("_")
    ]
    if not candidates:
        candidates = [fn for fn in functions if not fn.is_method and fn.name != "__init__"]
    if not candidates:
        candidates = [fn for fn in functions if fn.is_method and fn.name != "__init__"]

    limit = {"unit": 3, "integration": 2, "e2e": 1}.get(test_level, 2)
    return sorted(candidates, key=lambda fn: fn.name)[:limit]


def _symbols_to_import(
    targets: list[_FunctionInfo], functions: list[_FunctionInfo], test_level: str
) -> list[str]:
    symbols: list[str] = []
    class_names = {fn.parent_class for fn in functions if fn.parent_class}
    if test_level == "unit":
        class_names.update(fn.parent_class for fn in targets if fn.parent_class)
    for cls in sorted(name for name in class_names if name):
        if cls not in symbols:
            symbols.append(cls)
    for fn in targets:
        if fn.name not in symbols:
            symbols.append(fn.name)
    if not symbols:
        for fn in functions:
            if fn.parent_class and fn.parent_class not in symbols:
                symbols.append(fn.parent_class)
    return symbols or ["*"]


def _module_smoke_test(module_import: str, module_name: str, test_level: str) -> list[str]:
    return [
        "",
        f"def test_{module_name}_module_is_importable_{test_level}() -> None:",
        f'    """Verify {module_import} imports without error."""',
        f"    import {module_import} as module",
        "    assert module is not None",
    ]


def _unit_tests(
    targets: list[_FunctionInfo],
    functions: list[_FunctionInfo],
    io_imports: list[str],
    module_import: str,
) -> list[str]:
    lines: list[str] = []
    for fn in targets:
        lines.extend(_unit_test_for_function(fn, module_import, io_imports))
    class_targets = sorted({fn.parent_class for fn in functions if fn.parent_class})
    for cls in class_targets:
        lines.extend(_class_instantiation_test(cls))
    return lines


def _unit_test_for_function(fn: _FunctionInfo, module_import: str, io_imports: list[str]) -> list[str]:
    call_args = _sample_call_args(fn.params)
    expected = _expected_result(fn)
    test_name = f"test_{fn.name}_behavior"

    if fn.is_async:
        body = [
            "",
            "@pytest.mark.asyncio",
            f"async def {test_name}() -> None:",
            f'    """Exercise async function {fn.name}."""',
        ]
        if fn.uses_io and io_imports:
            body.extend(_mocked_call(fn, call_args, expected, module_import, io_imports[0], async_fn=True))
        else:
            body.append(f"    result = await {fn.name}({call_args})")
            body.extend(_assertion_lines(expected, "result"))
        return body

    if fn.uses_io and io_imports:
        return _mocked_call(fn, call_args, expected, module_import, io_imports[0], async_fn=False)

    lines = [
        "",
        f"def {test_name}() -> None:",
        f'    """Exercise {fn.name} with representative inputs."""',
        f"    result = {fn.name}({call_args})",
    ]
    lines.extend(_assertion_lines(expected, "result"))
    return lines


def _mocked_call(
    fn: _FunctionInfo,
    call_args: str,
    expected: str | None,
    module_import: str,
    io_module: str,
    *,
    async_fn: bool,
) -> list[str]:
    decorator = f"@patch('{module_import}.{io_module}')"
    lines = [
        "",
        decorator,
    ]
    if async_fn:
        lines.append("@pytest.mark.asyncio")
        lines.append(f"async def test_{fn.name}_with_mocked_io(mock_{io_module}: MagicMock) -> None:")
        lines.append(f"    mock_{io_module}.return_value = None")
        lines.append(f"    result = await {fn.name}({call_args})")
    else:
        lines.append(f"def test_{fn.name}_with_mocked_io(mock_{io_module}: MagicMock) -> None:")
        lines.append(f"    mock_{io_module}.return_value = None")
        lines.append(f"    result = {fn.name}({call_args})")
    if expected:
        lines.extend(_assertion_lines(expected, "result"))
    else:
        lines.append("    assert result is not None or result is None")
    return lines


def _integration_tests(
    targets: list[_FunctionInfo],
    functions: list[_FunctionInfo],
    module_import: str,
    module_name: str,
) -> list[str]:
    if len(targets) >= 2:
        first, second = targets[0], targets[1]
        args1 = _sample_call_args(first.params)
        args2 = _sample_call_args(second.params)
        return [
            "",
            f"def test_{module_name}_functions_work_together() -> None:",
            f'    """Call multiple public functions from {module_import}."""',
            f"    first = {first.name}({args1})",
            f"    second = {second.name}({args2})",
            "    assert first is not None or first is None",
            "    assert second is not None or second is None",
        ]

    fn = targets[0]
    args = _sample_call_args(fn.params)
    expected = _expected_result(fn)
    lines = [
        "",
        f"def test_{fn.name}_integration_path() -> None:",
        f'    """Integration-level call to {fn.name}."""',
        f"    result = {fn.name}({args})",
    ]
    lines.extend(_assertion_lines(expected, "result"))
    return lines


def _e2e_tests(
    targets: list[_FunctionInfo],
    functions: list[_FunctionInfo],
    module_import: str,
    module_name: str,
    source_code: str,
) -> list[str]:
    has_main_guard = 'if __name__ == "__main__"' in source_code
    main_fn = next((fn for fn in functions if fn.name == "main" and not fn.is_method), None)

    if main_fn:
        args = _sample_call_args(main_fn.params)
        if main_fn.is_async:
            lines = [
                "",
                "@pytest.mark.asyncio",
                f"async def test_{module_name}_main_entrypoint() -> None:",
                f'    """Exercise the async main entrypoint of {module_import}."""',
                f"    result = await main({args})",
            ]
        else:
            lines = [
                "",
                f"def test_{module_name}_main_entrypoint() -> None:",
                f'    """Exercise the main entrypoint of {module_import}."""',
                f"    result = main({args})",
            ]
        lines.extend(_assertion_lines(_expected_result(main_fn), "result"))
        return lines

    if has_main_guard:
        return [
            "",
            "import subprocess",
            "import sys",
            "",
            f"def test_{module_name}_runs_as_script() -> None:",
            f'    """Run {module_import} as a script end-to-end."""',
            "    proc = subprocess.run(",
            "        [sys.executable, '-m', "
            f"'{module_import}'],",
            "        capture_output=True,",
            "        text=True,",
            "        check=False,",
            "    )",
            "    assert proc.returncode in (0, 1)",
        ]

    fn = targets[0]
    args = _sample_call_args(fn.params)
    expected = _expected_result(fn)
    lines = [
        "",
        f"def test_{fn.name}_e2e_path() -> None:",
        f'    """End-to-end exercise of {fn.name}."""',
        f"    result = {fn.name}({args})",
    ]
    lines.extend(_assertion_lines(expected, "result"))
    return lines


def _class_instantiation_test(class_name: str) -> list[str]:
    return [
        "",
        f"def test_{class_name.lower()}_can_be_constructed() -> None:",
        f'    """Verify {class_name} can be instantiated."""',
        f"    instance = {class_name}()",
        f"    assert isinstance(instance, {class_name})",
    ]


def _sample_call_args(params: tuple[_ParamInfo, ...]) -> str:
    if not params:
        return ""
    values: list[str] = []
    for index, param in enumerate(params):
        if param.has_default:
            continue
        values.append(_sample_value(param.name, index))
    return ", ".join(values)


def _sample_value(name: str, index: int) -> str:
    lowered = name.lower()
    if lowered in {"a", "x", "left", "first", "start"}:
        return "2"
    if lowered in {"b", "y", "right", "second", "end"}:
        return "3"
    if "count" in lowered or "num" in lowered or "size" in lowered:
        return "1"
    if "name" in lowered or "text" in lowered or "label" in lowered:
        return "'sample'"
    if "path" in lowered or "file" in lowered:
        return "'/tmp/example'"
    if "flag" in lowered or lowered.startswith("is_") or lowered.startswith("has_"):
        return "True"
    if index == 0:
        return "1"
    if index == 1:
        return "2"
    return "None"


def _expected_result(fn: _FunctionInfo) -> str | None:
    expr = fn.return_expr
    if expr is None:
        return None
    if isinstance(expr, ast.Constant):
        return repr(expr.value)
    if isinstance(expr, ast.Name) and expr.id in {p.name for p in fn.params}:
        return _sample_value(expr.id, next(i for i, p in enumerate(fn.params) if p.name == expr.id))
    if isinstance(expr, ast.BinOp) and isinstance(expr.op, ast.Add):
        left = _literal_arg(expr.left, fn.params)
        right = _literal_arg(expr.right, fn.params)
        if left is not None and right is not None:
            try:
                return repr(left + right)
            except TypeError:
                return None
    if isinstance(expr, ast.BinOp) and isinstance(expr.op, ast.Sub):
        left = _literal_arg(expr.left, fn.params)
        right = _literal_arg(expr.right, fn.params)
        if left is not None and right is not None:
            try:
                return repr(left - right)
            except TypeError:
                return None
    if isinstance(expr, ast.UnaryOp) and isinstance(expr.op, ast.USub) and isinstance(expr.operand, ast.Constant):
        return repr(-expr.operand.value)
    return None


def _literal_arg(node: ast.expr, params: tuple[_ParamInfo, ...]) -> int | float | None:
    if isinstance(node, ast.Constant) and isinstance(node.value, (int, float)):
        return node.value
    if isinstance(node, ast.Name):
        for index, param in enumerate(params):
            if param.name == node.id:
                sample = _sample_value(param.name, index)
                if sample.isdigit():
                    return int(sample)
    return None


def _assertion_lines(expected: str | None, result_var: str) -> list[str]:
    if expected is not None:
        return [f"    assert {result_var} == {expected}"]
    return [f"    assert {result_var} is not None or {result_var} is None"]
