# Phase 2: Python Workflow Depth - Research

**Researched:** 2026-03-31
**Domain:** AST-guided Python test generation for CLI workflows
**Confidence:** MEDIUM-HIGH

## User Constraints (from CONTEXT.md)

### Locked Decisions
No `*-CONTEXT.md` was found for this phase. Locked decisions inferred from roadmap:
- AST-aware source analysis for better test generation.
- Structured test pyramid generation (`unit`, `integration`, `e2e` templates).
- Deterministic "fix failing generated test" loop.

### Claude's Discretion
- Minimal viable architecture for a small Python CLI.
- Library selection where roadmap does not force a specific tool.
- Test strategy and execution order.

### Deferred Ideas (OUT OF SCOPE)
- Enterprise policy controls (Phase 3+)
- Integrations/extensions (Phase 4+)
- Multi-step autonomous execution (Phase 5+)

## Summary

For this codebase, the fastest reliable Phase 2 path is to keep generation and orchestration in pure Python stdlib + pytest, and add structured components around the existing `gen-tests` command: an AST fact extractor, a test template renderer by pyramid level, and a subprocess-based red/green fix loop. This avoids premature dependency growth while directly targeting the roadmap metrics (`pass rate`, `rewrite regression`).

Python's `ast` module already provides enough primitives for MVP analysis (`NodeVisitor`, source segment extraction, child traversal) and can drive deterministic template selection from function signatures, return hints, async usage, raises, and side-effect heuristics. For deterministic correction, run pytest in a separate process each iteration and use machine-readable output (`--junit-xml`) plus fixed retry bounds; avoid repeated `pytest.main()` in-process reruns.

**Primary recommendation:** Implement Phase 2 with stdlib `ast` + pytest subprocess orchestration first; introduce LibCST only if Phase 2.2+ requires format-preserving source rewrites.

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Python `ast` (stdlib) | Python 3.10+ | Parse/analyze modules into test-generation facts | Zero dependency, official API, enough for analysis-centric generation |
| `pytest` | `>=8.0.0` (already present) | Execute generated tests and classify failures | Existing project standard and strong CLI/filter support |
| Python `subprocess` (stdlib) | Python 3.10+ | Deterministic isolated test-run loop | Official process isolation; avoids pytest re-entry pitfalls |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `junitparser` | latest | Parse pytest JUnit XML robustly | Prefer over ad-hoc XML parsing once loop is stable |
| `jinja2` | latest | Structured test template rendering | Use if string templates become hard to maintain |
| `libcst` | `1.8.x` | Lossless CST for precise rewrites | Only if fix loop must rewrite source while preserving formatting/comments |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| stdlib `ast` facts | `libcst` facts | Better fidelity but more complexity and install/build overhead |
| direct string templates | `jinja2` | Better maintainability vs extra dependency |
| parsing console output | JUnit XML parser | XML is more deterministic and less brittle |

**Installation:**
```bash
pip install -e ".[dev]"
pip install junitparser jinja2
```

## Architecture Patterns

### Recommended Project Structure
```text
src/ai_code_assistant/
├── services/
│   ├── ast_analyzer.py          # Source -> AnalysisFacts
│   ├── test_templates.py        # unit/integration/e2e template builders
│   ├── pyramid_generator.py     # Orchestrates templates by level
│   └── fix_loop.py              # Deterministic test->diagnose->regenerate
├── models/
│   └── analysis.py              # Typed dataclasses for AST facts/failures
└── cli.py                       # New flags, phase orchestration entrypoints
```

### Pattern 1: AST Facts Extraction (Read-Only Visitor)
**What:** Parse each source file once and produce normalized `AnalysisFacts` (functions/classes/imports/async/raises/io hints).
**When to use:** Before any generation step; cache by file hash.
**Example:**
```python
# Source: https://docs.python.org/3/library/ast.html
import ast
from dataclasses import dataclass

@dataclass
class FunctionFact:
    name: str
    args: list[str]
    is_async: bool
    has_return_annotation: bool

class FunctionCollector(ast.NodeVisitor):
    def __init__(self) -> None:
        self.functions: list[FunctionFact] = []

    def visit_FunctionDef(self, node: ast.FunctionDef) -> None:
        self.functions.append(
            FunctionFact(
                name=node.name,
                args=[a.arg for a in node.args.args],
                is_async=False,
                has_return_annotation=node.returns is not None,
            )
        )
        self.generic_visit(node)
```

### Pattern 2: Test Pyramid Template Router
**What:** Route each discovered symbol to one or more templates (`unit`, `integration`, `e2e`) based on deterministic rules.
**When to use:** During `gen-tests` generation before LLM fallback.
**Example routing rules:**
- `unit`: pure functions, no I/O/import side-effect hints.
- `integration`: module boundaries, DB/http/file interactions.
- `e2e`: CLI command flows (`subprocess`) and file-write behavior.

### Pattern 3: Deterministic Fix Loop
**What:** Bounded loop with fixed max iterations, isolated pytest invocation, structured failure parse, deterministic prompt context.
**When to use:** Only on generated test failures.
**Example command:**
```bash
python -m pytest -q --maxfail=1 --junit-xml .ai-code-assistant/last-run.xml -m "unit or integration or e2e"
```

### Anti-Patterns to Avoid
- **In-process rerun loop with `pytest.main()` repeatedly:** pytest docs warn repeated calls in same process may not reflect file changes due to import caching.
- **Regex-only Python parsing:** brittle for decorators, async defs, and nested constructs.
- **Unbounded self-healing loop:** can oscillate indefinitely and burn tokens/time.
- **Single mega-template:** harms pyramid separation and traceability.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Python parser | custom parser/token walker | stdlib `ast` | Correct grammar handling, maintained by CPython |
| Test runner lifecycle | bespoke in-process harness | `pytest` CLI via `subprocess.run` | Better isolation and deterministic exit handling |
| Structured failure format | regex over terminal text | `--junit-xml` + XML parser | Stable schema; less formatter drift |
| Format-preserving rewrites | homemade text patch heuristics | `libcst` (if needed) | Handles comments/whitespace/parentheses safely |

**Key insight:** custom parsing/rewrite logic becomes correctness debt quickly; keep MVP deterministic by delegating complex concerns to battle-tested tooling.

## Common Pitfalls

### Pitfall 1: AST Facts Too Shallow
**What goes wrong:** Generated tests ignore async/raises/type hints and become placeholder-heavy.
**Why it happens:** Collecting only names and argument lists.
**How to avoid:** Include return annotations, decorators, raise statements, called imports, and source segments.
**Warning signs:** High "generated but failing due to wrong call shape" rate.

### Pitfall 2: Pyramid Marker Drift
**What goes wrong:** Generated tests use unregistered or inconsistent markers.
**Why it happens:** Template strings diverge from `pyproject.toml`.
**How to avoid:** Centralize marker constants and validate against configured marker list.
**Warning signs:** `PytestUnknownMarkWarning` or marker typos.

### Pitfall 3: Non-Deterministic Fix Iterations
**What goes wrong:** Same inputs produce different fixes across retries.
**Why it happens:** Non-canonical prompt context, multiple failing tests at once, no fixed ordering.
**How to avoid:** `--maxfail=1`, stable seed/temperature, canonical failure payload, bounded retries (e.g., 3).
**Warning signs:** Flip-flopping edits and inconsistent patch diffs.

### Pitfall 4: Cache-Contaminated Loop Results
**What goes wrong:** Loop appears green/red due to stale pytest cache state.
**Why it happens:** Reusing cache across correction attempts.
**How to avoid:** pass `--cache-clear` in correction mode where reproducibility > speed.
**Warning signs:** Different results on immediate rerun without code changes.

## Code Examples

Verified patterns from official sources:

### Parse module and walk children
```python
# Source: https://docs.python.org/3/library/ast.html
import ast

tree = ast.parse(source_code)
for node in ast.walk(tree):
    if isinstance(node, ast.Call):
        ...
```

### Isolated deterministic pytest execution
```python
# Sources:
# - https://docs.python.org/3/library/subprocess.html
# - https://pytest.org/en/stable/reference/exit-codes.html
import subprocess

result = subprocess.run(
    ["python", "-m", "pytest", "-q", "--maxfail=1", "--junit-xml", "last-run.xml"],
    capture_output=True,
    text=True,
    timeout=120,
    check=False,
)

# 0=all pass, 1=test failures, 2..5 execution/usage/internal/no-tests conditions
if result.returncode == 1:
    handle_failing_test_xml("last-run.xml")
```

### Marker-based pyramid selection
```python
# Source: https://pytest.org/en/stable/example/markers.html
# pyproject.toml
[tool.pytest.ini_options]
markers = [
  "unit: fast isolated tests",
  "integration: cross-module behavior tests",
  "e2e: end-to-end CLI execution tests",
]
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Pure prompt-only generation | Hybrid structural (AST/CST) + prompt generation | 2023-2026 ecosystem trend | Better determinism and less hallucinated API usage |
| Free-form retry loops | Bounded, machine-parseable test-fix loops | Matured CI/testgen tooling era | Predictable runtime and lower oscillation risk |
| Regex parsing of failures | Structured outputs (JUnit/report logs) | Established in modern pytest usage | More robust failure classification |

**Deprecated/outdated:**
- Repeated `pytest.main()` loop in same process for dynamic code updates: pytest docs caution import caching can hide file changes between runs.

## Practical Implementation Checklist (MVP)

1. Add `AnalysisFacts` dataclasses and `ast_analyzer.py` for function/class/import/async/raise extraction.
2. Extend `gen-tests` with `--pyramid unit|integration|e2e|all` and default `all`.
3. Implement `test_templates.py` with three explicit template families and marker constants.
4. Add deterministic routing rules from AST facts to template type(s).
5. Write generated tests to stable paths (`tests/unit/...`, `tests/integration/...`, `tests/e2e/...`) or equivalent marker-based layout.
6. Implement `fix_loop.py`:
   - run pytest in subprocess (`--maxfail=1 --junit-xml ...`),
   - parse first failing case,
   - regenerate only impacted test file,
   - stop at `max_attempts=3`.
7. Add CLI flags: `--fix-failing-generated-tests`, `--max-fix-attempts`, `--pytest-args`.
8. Persist audit events for each fix attempt (`attempt`, `nodeid`, `returncode`, `edited_file`).
9. Add unit tests for AST extractor and template router, integration tests for fix loop with seeded failing fixture repo, and e2e tests for CLI flow.
10. Gate completion with reproducibility check: same input repo + same flags -> same generated outputs and same loop decision path.

## Minimal Test Strategy for This Phase

- **Unit tests (`@pytest.mark.unit`)**
  - AST extraction correctness (signatures, async, raises, imports).
  - Template rendering determinism (snapshot/string assertions).
  - Fix-loop state transitions (mocked subprocess and XML parser).

- **Integration tests (`@pytest.mark.integration`)**
  - End-to-end from source file -> generated test files -> one real pytest run.
  - Deterministic single-failure repair path on controlled fixture project.

- **E2E tests (`@pytest.mark.e2e`)**
  - CLI invocation via subprocess on temp repo.
  - JSON output and audit log invariants across repeated runs.

## Open Questions

1. **Should the fix loop edit source code, generated tests, or both?**
   - What we know: roadmap says "fix failing generated test loop", implying tests-first.
   - What's unclear: whether source rewrites are allowed in Phase 2.
   - Recommendation: constrain Phase 2 to generated tests only; defer source rewrites to later phase.

2. **Do we need LibCST in Phase 2?**
   - What we know: stdlib AST is enough for analysis and routing.
   - What's unclear: expected complexity of future code rewrite operations in this phase.
   - Recommendation: add integration seam (`AnalyzerProtocol`) now, postpone LibCST dependency until rewrite requirements appear.

## Sources

### Primary (HIGH confidence)
- Python `ast` docs: https://docs.python.org/3/library/ast.html (NodeVisitor/NodeTransformer/get_source_segment/walk)
- Python `subprocess` docs: https://docs.python.org/3/library/subprocess.html (`run`, `capture_output`, `timeout`, `check`)
- pytest markers docs: https://pytest.org/en/stable/example/markers.html (custom markers and selection)
- pytest exit codes docs: https://pytest.org/en/stable/reference/exit-codes.html
- pytest cache/re-run docs: https://pytest.org/en/stable/how-to/cache.html (`--lf`, `--ff`, `--cache-clear`, stepwise)
- pytest usage docs: https://pytest.org/en/stable/how-to/usage.html (`pytest.main()` caveat for repeated in-process runs)
- pytest output docs: https://pytest.org/en/stable/how-to/output.html (`--junit-xml`)

### Secondary (MEDIUM confidence)
- Jinja docs: https://jinja.palletsprojects.com/en/stable/ (templating option for structured generation)
- LibCST package/docs overview: https://pypi.org/project/libcst/ and https://libcst.readthedocs.io/en/latest/

### Tertiary (LOW confidence)
- Web search snippets on marker best practices and rerun plugin limitations (used only as directional; superseded by official pytest docs above).

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - primarily official Python and pytest docs aligned with current codebase.
- Architecture: MEDIUM-HIGH - practical extrapolation from official capabilities and current CLI structure.
- Pitfalls: MEDIUM-HIGH - documented pytest/python behavior plus common implementation failure modes.

**Research date:** 2026-03-31
**Valid until:** 2026-04-30
