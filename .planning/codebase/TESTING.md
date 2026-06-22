# Testing Patterns

**Analysis Date:** 2026-06-22

## Test Framework

**Runner:**
- **pytest** ≥8.0 — primary framework across Python subprojects
- Config: per-project `[tool.pytest.ini_options]` in `pyproject.toml` files
- **JUnit 5** — Kotlin/Gradle projects (`rabbit-mq/`, `modular-jvm-build/`)
- **Node.js built-in test runner** (`node:test`) — `dev-ex/tooling-adoption-tracker/`
- **Bazel Java tests** — `bazel-multibuild/java/src/test/java/com/example/myproject/TestApp.java`
- **Container structure tests** — `bazel-multibuild/java/container-structure-test.yaml` (not pytest)

**Assertion Library:**
- pytest built-in `assert`
- `unittest`-style via pytest
- Kotlin: `org.junit.jupiter.api.Assertions`
- TypeScript: `node:assert/strict`

**Run Commands:**
```bash
# Monorepo-wide (root Makefile)
make test                    # pytest -v --continue-on-collection-errors + ./gradlew test

# Monorepo CI (when Python changes)
pytest -q --cov=.            # .github/workflows/ci.yml

# Per-project (typical pattern)
cd ai-best-practices-examples/ai-code-assistant && pytest
cd rest-api-test-demo && pytest -v --cov=app --cov-report=term-missing
cd dev-ex/tooling-adoption-tracker && npm run build && npm test
cd modular-jvm-build && ./gradlew test
cd rabbit-mq && ./gradlew test
bazel test //...              # bazel-multibuild (when Bazel files change)
```

## Test File Organization

**Location:**
- Standard: top-level `tests/` directory sibling to `src/`
- Exception: `online-bookstore/test/` (singular), `projgen/src/tests/`
- Kotlin: `src/test/kotlin/` inside each Gradle module
- TypeScript: `tests/*.test.ts` with compiled output tested via `dist/tests/**/*.test.js`

**Naming:**
- Python: `test_<subject>.py` — e.g. `test_cli.py`, `test_compatibility_rules.py`
- Smoke scaffolds: `test_smoke.py` in all `ci-cd-pipelines/*/tests/`
- Kotlin: `<Class>Test.kt` — e.g. `MessageUtilsTest.kt`
- TypeScript: `<module>.test.ts`

**Structure:**
```
# Mature Python project (ai-code-assistant)
tests/
├── conftest.py
├── test_cli.py              # top-level / legacy
├── unit/
│   └── test_*_unit.py
├── integration/
│   └── test_cli_integration.py
└── e2e/
    └── test_cli_e2e.py

# Scaffold Python project (ci-cd-pipelines)
tests/
└── test_smoke.py            # 2 assertions on src/main.py

# Empty placeholder (dev-env)
tests/
├── unit/.keep
├── integration/.keep
└── e2e/.keep
```

## Test Structure

**Suite Organization (best-in-repo example — ai-code-assistant):**
```python
# tests/conftest.py — auto-markers by directory
def pytest_collection_modifyitems(config, items):
    for item in items:
        path = Path(str(item.fspath))
        if "e2e" in path.parts:
            item.add_marker(pytest.mark.e2e)
        elif "integration" in path.parts:
            item.add_marker(pytest.mark.integration)
        else:
            item.add_marker(pytest.mark.unit)
```

**Pytest markers (ai-code-assistant `pyproject.toml`):**
```toml
markers = [
  "unit: fast isolated tests for single modules/functions",
  "integration: cross-module behavior tests without full external stack",
  "e2e: end-to-end CLI execution tests via subprocess or real file writes",
]
```

**Patterns:**
- Setup: `@pytest.fixture` with `autouse=True` for state reset — e.g. `rest-api-test-demo/tests/conftest.py` clears in-memory store per test
- Teardown: yield fixtures or rely on autouse reset
- Assertion: plain `assert`; typed test functions with `-> None` return annotation

## Mocking

**Framework:** pytest `monkeypatch` fixture (no `unittest.mock` dependency required)

**Patterns:**
```python
# ai-code-assistant/tests/unit/test_llm_adapter_unit.py
def test_generate_prefers_openai_path_when_api_key_is_set(monkeypatch):
    adapter = LLMAdapter(api_key="token")
    def fake_openai(source_code, module_name, facts, test_level):
        return "def test_calc_model():\n    assert True\n"
    monkeypatch.setattr(adapter, "_generate_with_openai", fake_openai)
    content = adapter.generate_tests(...)
    assert "def test_calc_model" in content

# ai-code-assistant/tests/integration/test_cli_integration.py
def test_single_file_mode_calls_output_with_expected_target(tmp_path, monkeypatch):
    monkeypatch.chdir(tmp_path)
    monkeypatch.setattr(cli, "_output_one", fake_output)
    exit_code = cli.main(["gen-tests", str(source), "--dry-run"])
    assert exit_code == 0
```

**What to Mock:**
- LLM/external API adapters — `ai-best-practices-examples/ai-code-assistant/tests/unit/test_llm_adapter_unit.py`
- CLI side effects (file writes, subprocess) — `tests/integration/`, `tests/e2e/`
- ComfyUI live calls — gated behind env var, not mocked when enabled

**What NOT to Mock:**
- Pure functions under test — e.g. `dev-ex/inner-loop-friction-scorer/tests/test_composite.py` tests `compute_composite_score()` directly
- In-memory FastAPI apps — `rest-api-test-demo/tests/test_api.py` uses real `TestClient`

## Fixtures and Factories

**Test Data:**
```python
# knowledge-qa-system/tests/conftest.py — chroma/isolation fixtures
# knowledge-qa-system/tests/fixtures/sample_doc.txt
# knowledge-qa-system/tests/fixtures/sample_doc.pdf
# environment-drift-detector/tests/fixtures/manifests/sample-manifest.yaml
```

**Location:**
- `tests/fixtures/` for static files
- `tests/conftest.py` for shared pytest fixtures
- `tmp_path` built-in fixture for ephemeral filesystem tests

## Coverage

**Requirements:** No enforced coverage threshold repo-wide; optional per-project

**View Coverage:**
```bash
# rest-api-test-demo local CI
pytest -v --cov=app --cov-report=term-missing

# Monorepo CI
pytest -q --cov=.    # uploads via codecov-action@v5 in .github/workflows/ci.yml
```

**Coverage gaps:** Most scaffold projects (`ci-cd-pipelines/*`, `dev-env/*`) have no coverage configuration.

## Test Types

**Unit Tests:**
- Scope: single function/module, no I/O
- Location: `tests/unit/` or top-level `tests/test_*.py`
- Examples: `ai-best-practices-examples/ai-code-assistant/tests/unit/`, `dev-ex/inner-loop-friction-scorer/tests/test_composite.py`

**Integration Tests:**
- Scope: cross-module wiring, mocked externals
- Location: `tests/integration/`
- Example: `ai-best-practices-examples/ai-code-assistant/tests/integration/test_cli_integration.py`

**E2E Tests:**
- Scope: full CLI subprocess or real file writes
- Location: `tests/e2e/`
- Example: `ai-best-practices-examples/ai-code-assistant/tests/e2e/test_cli_e2e.py`

**Smoke Tests:**
- Scope: minimal "does it import and run" checks
- Pattern: import from `src/main.py`, assert one happy-path output
- All 6 `ci-cd-pipelines/*/tests/test_smoke.py` files follow this pattern (2 tests each, real logic not just `assert True`)

**Opt-in Live Smoke:**
- `ai-best-practices-examples/ai-image-video-generator/tests/test_smoke_comfyui.py` — skipped unless `AIVG_RUN_SMOKE=1`

## Common Patterns

**Async Testing:**
- FastAPI: synchronous `TestClient` wrapper — e.g. `online-bookstore/test/test_main.py`, `rest-api-test-demo/tests/test_api.py`
- No `pytest-asyncio` config detected in pyproject files

**Error Testing:**
```python
# platform-audit-template/scripts/test_gcp_billing_summary.py
with pytest.raises(ValueError):
    datetime.strptime("invalid", "%Y-%m-%d")

# online-bookstore/test/test_main.py
response = client.get("/v1/books/9999")
assert response.status_code == 404
```

**Legacy sys.path pattern (avoid in new tests):**
```python
# online-bookstore/test/test_main.py, ci-cd-pipelines/*/tests/test_smoke.py
import sys
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'src'))
```

Prefer `[tool.pytest.ini_options] pythonpath = ["src"]` in `pyproject.toml`.

---

## Build Verification Status by Subproject

Legend:
- **Real tests** — multiple meaningful assertions, covers business logic
- **Smoke-only** — minimal scaffold tests or single-file smoke
- **Placeholder** — `.keep` files only, no executable tests
- **None** — no test directory or files

### ai-best-practices-examples/

| Subproject | Test files | Type | CI | Notes |
|---|---:|---|---|---|
| `ai-code-assistant` | 15+ pytest | **Real tests** | Monorepo Python CI | unit/integration/e2e pyramid; `conftest.py` auto-markers |
| `ai-image-video-generator` | 9 pytest | **Real tests** | Monorepo Python CI | Includes opt-in ComfyUI live smoke |
| `chat-ai` | 5 pytest | **Real tests** | Monorepo Python CI | API, memory, tools, UI scaffold |
| `domain-expert-ai` | 6 pytest | **Real tests** | Monorepo Python CI | Training, eval, guardrails, data curation |
| `knowledge-qa-system` | 3 pytest | **Real tests** | Monorepo Python CI | Ingest, chunking, health; chroma fixtures |

### c0de-quality-and-analysis/

| Subproject | Test files | Type | CI | Notes |
|---|---:|---|---|---|
| `api-breaking-change-detector` | 1 pytest | **Real tests** | Local + monorepo | `test_compatibility_rules.py`; CI runs compat script |
| `dead-code-surface-reporter` | 1 pytest | **Real tests** | Local + monorepo | Prioritizer logic |
| `kotlin-custom-detekt-rules-library` | 1 Kotlin | **Real tests** | Local Gradle CI | CustomRulesTest.kt |
| `license-compliance-scanner` | 1 pytest | **Real tests** | Local + monorepo | Policy engine |
| `security-hotspot-annotator` | 1 pytest | **Real tests** | Local + monorepo | Comment formatter |

### ci-cd-pipelines/

| Subproject | Test files | Type | CI | Notes |
|---|---:|---|---|---|
| `canary-deployment-controller` | 1 pytest | **Smoke-only** | Monorepo Python CI | `test_smoke.py` — 2 routing assertions |
| `flaky-pipeline-gate` | 1 pytest | **Smoke-only** | Monorepo Python CI | `test_smoke.py` — gate decision logic (actually non-trivial) |
| `pipeline-cost-analyzer` | 1 pytest | **Smoke-only** | Monorepo Python CI | `test_smoke.py` |
| `pipeline-telemetry-exporter` | 1 pytest | **Smoke-only** | Monorepo Python CI | `test_smoke.py` |
| `release-lead-time-calculator` | 1 pytest | **Smoke-only** | Monorepo Python CI | `test_smoke.py` |
| `self-service-pipeline-template-engine` | 1 pytest | **Smoke-only** | Monorepo Python CI | `test_smoke.py` |

Note: ci-cd-pipelines smoke tests exercise real `src/main.py` functions (not empty `pass`), but each project has only one test file with ~2 test functions — treat as scaffolds needing expansion.

### dev-env/

| Subproject | Test files | Type | CI | Notes |
|---|---:|---|---|---|
| `devcontainer-feature-library` | 0 | **Placeholder** | Monorepo only | `tests/unit/.keep`, `tests/integration/.keep`, `tests/e2e/.keep` |
| `environment-drift-detector` | 0 (+fixtures) | **Placeholder** | Monorepo only | Fixtures in `tests/fixtures/` but no test functions; `.keep` in unit/integration |
| `local-service-mesh` | 0 | **Placeholder** | Monorepo only | `tests/e2e/.keep`, `tests/integration/.keep` |
| `onboarding-automation-cli` | 0 | **Placeholder** | Monorepo only | Kotlin project; `tests/unit/.keep`, `tests/integration/.keep` |
| `remote-dev-environment-orchestrator` | 0 | **Placeholder** | Monorepo only | `pyproject.toml` present; empty `tests/` with `.keep` files |

### dev-ex/

| Subproject | Test files | Type | CI | Notes |
|---|---:|---|---|---|
| `developer-satisfaction-pulse-system` | 2 pytest | **Real tests** | Monorepo Python CI | Health + scoring |
| `inner-loop-friction-scorer` | 2 pytest | **Real tests** | Monorepo Python CI | Composite score + main |
| `platform-changelog-migration-generator` | 2 pytest | **Real tests** | Monorepo Python CI | Diff engine + Python rewriter |
| `tooling-adoption-tracker` | 1 TS | **Real tests** | **No CI job** | `funnelCalculator.test.ts`; requires `npm run build` before `npm test` |

### Standalone / demo projects

| Subproject | Test files | Type | CI | Notes |
|---|---:|---|---|---|
| `bazel-multibuild` | 1 Java + container test | **Real tests** | Monorepo Bazel CI | Bazel test target; not pytest |
| `modular-jvm-build` | 1 Kotlin | **Real tests** | Monorepo Gradle CI | `HealthControllerTest.kt` |
| `online-bookstore` | 1 pytest (6 tests) | **Real tests** | Monorepo Python CI | Uses `test/` not `tests/`; sys.path hack |
| `otel-demo-stack` | 1 pytest | **Real tests** | Local workflow (no pytest in CI) | CI runs curl health checks only |
| `projgen` | 2 pytest | **Real tests** | Monorepo Python CI | Under `projgen/src/tests/` |
| `rabbit-mq` | 2 Kotlin | **Real tests** | Monorepo Gradle CI | Gradle init-generated tests |
| `rest-api-test-demo` | 2 pytest | **Real tests** | **Dedicated CI** | pytest + coverage + Docker build |
| `workflow-api-demo` | 1 pytest | **Real tests** | Monorepo Python CI | API tests only; worker untested |
| `ws-chat-fast` | 0 | **None** | Monorepo Python CI | No test directory |
| `forgex` | 0 | **None** | None | Large README; no tests |
| `platform-audit-template` | 1 pytest | **Real tests** | Monorepo Python CI | Script tests in `scripts/test_gcp_billing_summary.py` |
| `tools/gradle_to_bazel` | 0 | **Placeholder** | Monorepo only | Empty `tests/` directory |

---

## CI vs Test Execution Matrix

| Has dedicated subproject CI | Runs pytest in CI | Has real locally runnable tests |
|---|---|---|
| `c0de-quality-and-analysis/*` (5) | Partial (scripts, not always pytest) | Yes |
| `rest-api-test-demo` | Yes | Yes |
| `otel-demo-stack` | No (health curl only) | Yes (local pytest) |
| All others | Monorepo path-filter CI only | Varies (see table above) |

### Monorepo CI behavior (`.github/workflows/ci.yml`)

- **Python job:** Installs ruff, mypy, pytest, coverage; runs `pytest -q --cov=.` on entire repo when any `.py` file changes
- **Kotlin job:** `./gradlew ktlintCheck detekt test` when `.kt` or Gradle files change
- **Bazel job:** `bazel build //...` (build only, not `bazel test`) when Bazel files change
- **Implication:** Python tests run in CI but may fail silently in local `make test` (`--continue-on-collection-errors`); Bazel tests are not executed in CI

### Projects with CI but weak test execution

| Project | CI runs | Tests actually executed in CI |
|---|---|---|
| `otel-demo-stack` | curl health + docker compose build | No pytest |
| `bazel-multibuild` | `bazel build //...` | No `bazel test` |
| `dev-ex/tooling-adoption-tracker` | Nothing project-specific | npm test not wired |
| `dev-env/*` | Monorepo pytest (collects nothing) | Zero tests |
| `ws-chat-fast`, `forgex` | Monorepo pytest (collects nothing) | Zero tests |

---

## Recommendations for New Tests

**Add tests here:**
- Python feature: `<subproject>/tests/test_<feature>.py`
- Python unit layer: `<subproject>/tests/unit/test_<module>_unit.py`
- Kotlin: `<module>/src/test/kotlin/<package>/<Class>Test.kt`
- TypeScript: `<subproject>/tests/<module>.test.ts`

**Follow ai-code-assistant patterns when building non-trivial Python CLIs:**
1. Add `[tool.pytest.ini_options]` with `pythonpath = ["src"]` to `pyproject.toml`
2. Create `tests/conftest.py` with directory-based markers
3. Split unit / integration / e2e under `tests/`
4. Use `monkeypatch` for external boundaries; `tmp_path` for filesystem

**For scaffolds (ci-cd-pipelines, dev-env):** Expand beyond `test_smoke.py` or replace `.keep` placeholders with at least one unit test per `src/` module before marking project "stable".

**Gitignore before testing:** Ensure `__pycache__/`, `.pytest_cache/`, `.venv/` are ignored — see `CONVENTIONS.md` git hygiene section. Tracked bytecode in `online-bookstore/test/__pycache__/` causes unnecessary CI noise.

---

*Testing analysis: 2026-06-22*
