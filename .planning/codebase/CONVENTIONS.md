# Coding Conventions

**Analysis Date:** 2026-06-22

## Naming Patterns

**Files:**
- Python modules: `snake_case.py` — e.g. `repo_scanner.py`, `llm_adapter.py` in `ai-best-practices-examples/ai-code-assistant/src/ai_code_assistant/`
- Python tests: `test_<module>.py` or `test_<feature>.py` — e.g. `tests/test_cli.py`, `tests/unit/test_llm_adapter_unit.py`
- Kotlin sources: PascalCase class files — e.g. `MessageUtilsTest.kt`, `HealthControllerTest.kt`
- TypeScript tests: `<module>.test.ts` co-located under `tests/` — e.g. `dev-ex/tooling-adoption-tracker/tests/funnelCalculator.test.ts`
- CI/CD pipeline scaffolds: each subproject under `ci-cd-pipelines/<name>/` with `src/main.py` entry and `tests/test_smoke.py`

**Functions:**
- Python: `snake_case` with explicit return type hints in newer projects — e.g. `build_parser()`, `compute_composite_score()` in `dev-ex/inner-loop-friction-scorer/src/friction_scorer/core/composite.py`
- Kotlin: camelCase test methods — e.g. `testGetMessage()` in `rabbit-mq/app/src/test/kotlin/org/example/app/MessageUtilsTest.kt`
- CLI entry points: `main()` returning exit code in `ai-best-practices-examples/ai-code-assistant/src/ai_code_assistant/cli.py`

**Variables:**
- Python: `snake_case`; use typed `dict[str, ...]` and `list[...]` in Python 3.10+ projects
- Constants: `UPPER_SNAKE_CASE` where used — e.g. env var names referenced in `ai-best-practices-examples/knowledge-qa-system/src/app/ingestion/adapters/notion_adapter.py`

**Types:**
- Pydantic models in FastAPI services — e.g. `dev-ex/developer-satisfaction-pulse-system/src/app/`
- `@dataclass` and plain dicts in CLI/scaffold projects — e.g. `ci-cd-pipelines/flaky-pipeline-gate/src/main.py`
- Kotlin data classes and JUnit 5 test classes

## Code Style

**Formatting:**
- Root `Makefile` target `lint` runs `ruff .` and `ktlintCheck` — see `Makefile`
- Root CI (`.github/workflows/ci.yml`) runs `ruff check .` and `mypy .` on all Python when `**/*.py` changes
- No repo-wide `ruff.toml`, `.prettierrc`, or `black` config detected; Ruff and mypy use defaults
- Kotlin CI runs `./gradlew ktlintCheck detekt test` when `**/*.kt` or `build.gradle*` changes

**Linting:**
- Python: Ruff (lint) + mypy (types) enforced at monorepo CI level
- Kotlin: ktlint + detekt via Gradle — e.g. `c0de-quality-and-analysis/kotlin-custom-detekt-rules-library/config/detekt-custom-rules.yml`
- TypeScript: `tsc -p tsconfig.json` compile check in `dev-ex/tooling-adoption-tracker/package.json`; no ESLint config detected

**Python version targets:**
- `>=3.10`: `ai-best-practices-examples/ai-code-assistant/pyproject.toml`, `projgen/pyproject.toml`
- `>=3.11`: `ai-best-practices-examples/chat-ai/pyproject.toml`, `dev-ex/developer-satisfaction-pulse-system/pyproject.toml`
- CI pins Python 3.12 in `.github/workflows/ci.yml`

## Import Organization

**Order (canonical pattern in newer Python projects):**
1. Standard library — e.g. `from pathlib import Path`, `import argparse`
2. Third-party — e.g. `import pytest`, `from fastapi import FastAPI`
3. Local/package imports — e.g. `from ai_code_assistant.adapters.llm_adapter import LLMAdapter`

**Path Aliases:**
- Setuptools `package-dir = {"" = "src"}` with `pythonpath = ["src"]` in pytest config — used in `ai-best-practices-examples/ai-code-assistant/pyproject.toml`, `ai-best-practices-examples/chat-ai/pyproject.toml`
- Legacy pattern: manual `sys.path.insert` or `sys.path.append` in tests — e.g. `online-bookstore/test/test_main.py`, `ci-cd-pipelines/flaky-pipeline-gate/tests/test_smoke.py`. Prefer `pyproject.toml` `[tool.pytest.ini_options] pythonpath` for new code.

## Error Handling

**Patterns:**
- Input validation: raise `ValueError` with descriptive messages — dominant pattern in `ai-best-practices-examples/domain-expert-ai/`, `ai-best-practices-examples/knowledge-qa-system/`, `ai-best-practices-examples/ai-code-assistant/src/ai_code_assistant/extensions.py`
- CLI errors: catch and return non-zero exit codes from `main()` — e.g. `ai-best-practices-examples/ai-code-assistant/src/ai_code_assistant/cli.py`
- External service fallbacks: broad `except Exception` with degraded fallback — e.g. `ai-best-practices-examples/ai-code-assistant/src/ai_code_assistant/adapters/llm_adapter.py`, `ai-best-practices-examples/chat-ai/src/ai_app/tools/weather.py`
- Ingestion pipelines: catch per-item failures, continue batch — e.g. `ai-best-practices-examples/knowledge-qa-system/src/app/domain/services/ingestion_service.py` uses `# noqa: BLE001` on broad catches

**Do this for new code:** Prefer explicit `ValueError`/`TypeError` at boundaries; avoid bare `except:`; document intentional broad catches.

## Logging

**Framework:** stdlib `logging` in FastAPI apps — e.g. `ws-chat-fast/app/chat.py` uses `logging.getLogger("uvicorn")`

**Patterns:**
- Module-level loggers in WebSocket/FastAPI handlers
- Audit trails as JSONL files in CLI tools — e.g. `ai-best-practices-examples/ai-code-assistant/src/ai_code_assistant/audit.py` writes to `.ai-code-assistant/audit.log.jsonl`
- No centralized structured logging library (structlog, loguru) detected

## Comments

**When to Comment:**
- Docstrings on public test helpers and non-obvious fixtures — e.g. `rest-api-test-demo/tests/conftest.py`
- Inline comments for CI/smoke opt-in behavior — e.g. `ai-best-practices-examples/ai-image-video-generator/tests/test_smoke_comfyui.py`
- Design rationale in README and `docs/` rather than in source

**JSDoc/TSDoc:**
- Not used; TypeScript relies on type signatures — e.g. `dev-ex/tooling-adoption-tracker/src/analytics/funnelCalculator.ts`

## Function Design

**Size:** CLI modules (`cli.py`) are large orchestrators; business logic kept in `services/`, `core/`, or `adapters/` subpackages

**Parameters:** Prefer explicit keyword args and typed dicts; CLI uses `argparse` subparsers — pattern in `ai-best-practices-examples/ai-code-assistant/src/ai_code_assistant/cli.py`

**Return Values:** Typed return annotations (`-> float`, `-> dict[str, str]`) in newer Python; Kotlin tests use JUnit assertions

## Module Design

**Exports:** Package `__init__.py` files mostly empty or minimal — e.g. `ai-best-practices-examples/chat-ai/src/ai_app/__init__.py`

**Barrel Files:** Not used; import from concrete modules directly

**Layout convention for Python subprojects:**
```
<subproject>/
├── pyproject.toml
├── README.md
├── src/<package_name>/     # implementation
└── tests/                  # pytest (sometimes unit/, integration/, e2e/)
```

**Layout convention for Gradle/Kotlin subprojects:**
```
<subproject>/
├── build.gradle.kts
├── README.md
└── <module>/src/main/kotlin/
    └── <module>/src/test/kotlin/
```

## Portfolio README Conventions

README quality varies widely across subprojects. Use this matrix when adding or updating docs.

| Subproject | README | Lines | Quality | Notes |
|---|---|---:|---|---|
| Root | `README.md` | 34 | Stale badges | Contains `yourusername` in CI/CodeQL/Release badge URLs |
| `PORTFOLIO.md` | Present | 13 | Index only | Competency mapping table; not a setup guide |
| `ai-best-practices-examples/ai-code-assistant` | Present | 78 | Good | Setup, usage, policy docs |
| `ai-best-practices-examples/ai-image-video-generator` | Present | 140 | Good | Notes ComfyUI workflow placeholders |
| `ai-best-practices-examples/chat-ai` | Present | 27 | Minimal | Setup present; thin feature docs |
| `ai-best-practices-examples/domain-expert-ai` | Present | 192 | Good | Training/eval pipeline documented |
| `ai-best-practices-examples/knowledge-qa-system` | Present | 25 | Minimal | Setup only |
| `c0de-quality-and-analysis/` (parent) | Present | 19 | Index | Lists 5 scaffolds; no run instructions |
| `c0de-quality-and-analysis/*` (5 tools) | Present | ~41 each | Scaffold | Consistent template; setup sections |
| `ci-cd-pipelines/` (parent) | Present | 13 | Index | Describes 6 pipeline concepts |
| `ci-cd-pipelines/*` (6 tools) | Present | ~40 each | Scaffold | Uniform README template |
| `dev-env/` (parent) | Present | 11 | Index | Points to 5 subprojects |
| `dev-env/*` (5 tools) | Present | 44–55 | Scaffold | Setup sections; tests are `.keep` stubs |
| `dev-ex/` (parent) | Present | 9 | Index | Points to 4 subprojects |
| `dev-ex/*` (4 tools) | Present | 46–51 | Scaffold | Setup sections |
| `bazel-multibuild` | Present | 4 | Placeholder | `//TODO: Add README content` |
| `modular-jvm-build` | Present | 65 | Good | Module layout, run instructions |
| `otel-demo-stack` | Present | 41 | Good | Docker/OTel setup |
| `rest-api-test-demo` | Present | 56 | Good | Test plan reference |
| `workflow-api-demo` | Present | 66 | Adequate | Architecture; weak setup section |
| `ws-chat-fast` | Present | 34 | Adequate | Structure/features; no install steps |
| `online-bookstore` | Present | 185 | Good | Detailed API docs; no quick-start |
| `projgen` | Present | 176 | Good | Mentions description placeholder in generated output |
| `forgex` | Present | 958 | Extensive | Large spec doc; uses "placeholder" in template context |
| `rabbit-mq` | Present | 4 | Placeholder | `//TODO: Add a description` |
| `platform-audit-template` | Present | 36 | Adequate | Audit script docs |

**README conventions to follow:**
- Include Problem / Solution / How to run sections — best examples: `modular-jvm-build/README.md`, `rest-api-test-demo/README.md`
- Parent category READMEs (`ci-cd-pipelines/README.md`, `c0de-quality-and-analysis/README.md`) are index-only; leaf projects carry setup instructions
- Replace `yourusername` in root badge URLs before publishing — `README.md` lines 3–5
- Avoid `//TODO: Add README` stubs — fix targets: `bazel-multibuild/README.md`, `rabbit-mq/README.md`

## Portfolio CI Conventions

### Monorepo-level workflows (`.github/workflows/`)

| Workflow | File | Scope |
|---|---|---|
| CI (Python/Kotlin/Bazel/Docs) | `.github/workflows/ci.yml` | Path-filtered: ruff, mypy, pytest, codecov; Gradle ktlint/detekt/test; Bazel build; mkdocs |
| CodeQL | `.github/workflows/codeql.yml` | Security scanning |
| Release Please | `.github/workflows/release-please.yml` | Release automation |
| GitHub Pages | `.github/workflows/pages.yml` | Static site deploy |
| README drift | `.github/workflows/readme-drift.yml` | Validates generated projects table |
| Static Pages (legacy) | `.github/workflows/static.yml` | Deploys entire repo to Pages |

### Subproject-local CI workflows

| Subproject | Workflow | What it runs |
|---|---|---|
| `c0de-quality-and-analysis/api-breaking-change-detector` | `.github/workflows/ci.yml` | `./scripts/check-api-compat.sh` |
| `c0de-quality-and-analysis/dead-code-surface-reporter` | `.github/workflows/ci.yml` | Project-specific check script |
| `c0de-quality-and-analysis/kotlin-custom-detekt-rules-library` | `.github/workflows/ci.yml` | Gradle build/test |
| `c0de-quality-and-analysis/license-compliance-scanner` | `.github/workflows/ci.yml` | Project-specific check script |
| `c0de-quality-and-analysis/security-hotspot-annotator` | `.github/workflows/ci.yml` | Project-specific check script |
| `rest-api-test-demo` | `.github/workflows/build.yml` | pytest with coverage + Docker build |
| `otel-demo-stack` | `.github/workflows/build.yml` | Health curl checks + docker compose build (no pytest in CI) |

### Subprojects without dedicated CI workflows

All other leaf projects rely solely on monorepo `.github/workflows/ci.yml` path filters, or have **no automated verification**:

- `ai-best-practices-examples/*` (5 projects) — covered by monorepo Python CI when `.py` changes
- `ci-cd-pipelines/*` (6 projects) — monorepo Python CI only
- `dev-env/*` (5 projects) — monorepo CI; no local workflow
- `dev-ex/*` (4 projects) — monorepo CI; TypeScript project has npm test script but no CI job
- `ws-chat-fast`, `online-bookstore`, `projgen`, `forgex`, `rabbit-mq`, `bazel-multibuild`, `modular-jvm-build`, `workflow-api-demo`, `platform-audit-template` — no local workflow

**Local dev verification:** Root `Makefile` targets `lint`, `test`, `docs`. The `test` target runs `pytest` repo-wide with `--continue-on-collection-errors` and `./gradlew test`, explicitly noting some failures are expected.

## Git Hygiene and Ignore Conventions

### Root `.gitignore` coverage

File: `.gitignore` — covers `.gradle`, `build/`, `.idea/` partials, `.DS_Store`, `.vscode/`, Kotlin/Eclipse/NetBeans artifacts.

### Missing from root `.gitignore` (should be added)

| Pattern | Present on disk | Tracked in git | Risk |
|---|---|---|---|
| `__pycache__/` | Yes, widespread | **Yes — ~915 entries** | Bytecode churn, noisy diffs |
| `*.pyc` | Yes | Tracked under `online-bookstore/`, `otel-demo-stack/`, etc. | Same |
| `.pytest_cache/` | Yes (`ci-cd-pipelines/.pytest_cache/`) | Partial | Cache leakage |
| `.venv/` | Yes (`projgen/.venv/`, `ai-best-practices-examples/chat-ai/.venv/`) | **Yes — ~2000+ entries in `projgen/.venv/`** | Committed virtualenv |
| `*.egg-info/` | Yes | Untracked noise in git status | Packaging artifacts |
| `.idea/` | Yes | **Yes — 6 files tracked** | IDE settings |
| `chroma_data/` | Yes in chat-ai, knowledge-qa | Check before commit | Local vector DB data |
| `.ai-code-assistant/` | Audit logs, checkpoints | Untracked in git status | Runtime artifacts |

### Subproject `.gitignore` files

Only 7 detected: root, `rabbit-mq/.gitignore`, `dev-env/*/.gitignore` (4 projects). Most Python subprojects inherit root ignore rules only — insufficient for Python-specific artifacts.

`rabbit-mq/.gitignore` correctly includes `.pytest_cache` — use as template for Python subprojects.

### Recommended ignore additions (root `.gitignore`)

```
__pycache__/
*.py[cod]
.pytest_cache/
.venv/
*.egg-info/
.mypy_cache/
.ruff_cache/
chroma_data/
.ai-code-assistant/
```

### Cleanup priority

1. Remove tracked `projgen/.venv/` from git index (largest offender)
2. Remove tracked `__pycache__/` and `*.pyc` under `online-bookstore/`, `otel-demo-stack/`, `rest-api-test-demo/`
3. Remove tracked `.idea/` files or expand `.idea/` ignore
4. Add patterns above to root `.gitignore`

---

*Convention analysis: 2026-06-22*
