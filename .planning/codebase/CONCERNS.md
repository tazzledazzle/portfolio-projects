# Codebase Concerns

**Analysis Date:** 2026-06-22

## Prioritized Portfolio Remediation

Ranked by impact for portfolio credibility and maintainability. Build verification was run on 2026-06-22 from the repo root and per-subproject where `Makefile`, `gradlew`, or `pyproject.toml` exist.

| Priority | Action | Impact | Effort |
|----------|--------|--------|--------|
| P0 | Remove committed virtualenv and bytecode from git (`projgen/.venv/`, `__pycache__/`, `.pytest_cache/`) and expand root `.gitignore` | Repo bloat (~2,000+ tracked venv files), clone noise, accidental secret risk | Medium |
| P0 | Fix `online-bookstore/src/main.py` broken FastAPI imports | Listed as **stable** in `portfolio.yaml` but tests fail on collection | Low |
| P0 | Fix `bazel-multibuild/java/BUILD` undefined rules (`oci_image`, `tar`, `container_structure_test`) | Listed as **beta**; `bazel test //...` fails; CI bazel job will fail when triggered | Medium |
| P1 | Expand `portfolio.yaml` + regenerate README table to index all project suites (40+ subprojects across 6 suites) | Visitors see only 5 of ~45 projects; large untracked work invisible | Medium |
| P1 | Replace `yourusername` badge URLs in `README.md` with real GitHub org/user | Broken CI/Release/CodeQL badges on every page view | Low |
| P1 | Fix root CI test strategy: `pytest -q --cov=.` from repo root yields **36 collection errors** (import/path collisions) | CI python job fails on any `.py` change repo-wide | High |
| P1 | Add `gradlew` + `repositories {}` to `dev-env/onboarding-automation-cli/` | `./gradlew test` impossible; system `gradle test` fails: "no repositories are defined" | Low |
| P1 | Add Gradle build files to `c0de-quality-and-analysis/kotlin-custom-detekt-rules-library/` | Subproject CI (`.github/workflows/ci.yml`) runs `./gradlew test detekt` but no `gradlew` or `build.gradle*` exists | Medium |
| P2 | Populate empty `pyproject.toml` in `dev-env/environment-drift-detector/` and `dev-env/remote-dev-environment-orchestrator/` | Projects have source + README but no installable package, no tests discovered | Medium |
| P2 | Add `pip install -e .` step (or document it) for `ai-best-practices-examples/domain-expert-ai/` | Tests fail without editable install (`ModuleNotFoundError: domain_expert_ai`) | Low |
| P2 | Reconcile design-doc inventory in `README.md` Legacy Design Documents section | Lists 11 filenames including `dd-advanced-logging-and-tracing.md` but only **10** files exist under `src/main/resources/` | Low |
| P2 | Refresh `PORTFOLIO.md` competency map | References non-existent `observability-stack`, broken anchor `README.md#portfolio-competency-mapping`, omits 6 project suites | Low |
| P2 | Populate `SECURITY.md` (currently empty; `PATCH_IMPLEMENTATION_SUMMARY.md` claims it was completed) | Security policy gap for open-source portfolio | Low |
| P3 | Add `gradlew` wrapper to `modular-jvm-build/` | Builds with system Gradle only; no self-contained `./gradlew test` | Low |
| P3 | Stop committing generated MkDocs output in `site/` (34 tracked HTML/CSS/JS files) | Docs drift; `mkdocs build` regenerates anyway | Low |
| P3 | Add tests to `ws-chat-fast/` or downgrade status from **stable** | Zero pytest files; root `make test` reports "no tests ran" | Medium |
| P3 | Decide fate of `forgex/` (956-line README spec, zero implementation) | Misleading as a "project"; overlaps conceptually with `projgen/` | Decision |

---

## Tech Debt

### Committed build artifacts and caches

- Issue: Root `.gitignore` is Gradle/IDE-focused only (`/.gitignore`). Python, Node, and pytest artifacts are not ignored repo-wide.
- Files: `projgen/.venv/` (**2,010 tracked files**), `online-bookstore/src/__pycache__/`, `online-bookstore/test/__pycache__/`, `otel-demo-stack/api/__pycache__/`, `platform-audit-template/scripts/__pycache__/`, `ci-cd-pipelines/.pytest_cache/`, `dev-env/remote-dev-environment-orchestrator/.pytest_cache/`
- Impact: Bloated clones, noisy diffs, risk of committing secrets into venv paths, reviewer fatigue.
- Fix approach: Add standard Python/Node ignores at root; `git rm -r --cached` on committed artifacts; enforce via pre-commit or CI grep check.

### Portfolio index covers 5 of ~45 projects

- Issue: `portfolio.yaml`, `README.md` project table, and `mkdocs.yml` nav all enumerate the same 5 legacy projects. Six newer suites with dozens of subprojects are absent.
- Files: `portfolio.yaml`, `README.md`, `mkdocs.yml`, `docs/design-docs/*/design.md` (only 5 design docs)
- Impact: New work in `ai-best-practices-examples/`, `ci-cd-pipelines/`, `c0de-quality-and-analysis/`, `dev-env/`, `dev-ex/`, `otel-demo-stack/`, `rest-api-test-demo/`, `workflow-api-demo/`, `modular-jvm-build/`, `platform-audit-template/`, `observability/` is invisible to readers.
- Fix approach: Extend `portfolio.yaml` with suite groupings or per-subproject entries; run `scripts/gen_readme_table.py`; add mkdocs nav sections per suite.

### Root pytest is not monorepo-safe

- Issue: `.github/workflows/ci.yml` runs `pytest -q --cov=.` from repo root. On 2026-06-22 this produced **70 passed, 36 errors** (collection/import failures).
- Files: `.github/workflows/ci.yml`, `Makefile` (`test` target also runs root pytest)
- Failing patterns:
  - `online-bookstore/test/test_main.py` — broken import in `src/main.py`
  - `ci-cd-pipelines/*/tests/` — pass in isolation (`make test` per subdir) but fail from root (package path issues)
  - `rest-api-test-demo/tests` — `ImportPathMismatchError` when collected with other `tests/` trees
  - `workflow-api-demo/api/tests/`, `otel-demo-stack/api/tests/` — similar path collisions
- Fix approach: Per-project pytest in matrix, `pytest.ini` with `testpaths`, or `tox`/`nox` orchestration; never collect all `tests/` trees simultaneously without package isolation.

### Root mypy is not monorepo-safe

- Issue: CI runs `mypy .` from root. Fails immediately on duplicate `conftest` module names across `ai-best-practices-examples/*/tests/`.
- Files: `.github/workflows/ci.yml`, `ai-best-practices-examples/ai-code-assistant/tests/conftest.py`, `ai-best-practices-examples/ai-image-video-generator/tests/conftest.py`
- Fix approach: Per-package `mypy` with `--explicit-package-bases`, or exclude `tests/` from root mypy.

### Bazel multibuild container targets broken

- Issue: `bazel-multibuild/java/BUILD` references `oci_image`, `tar`, and `container_structure_test` without loading the rules that define them. `bazel test //...` from `bazel-multibuild/java/` fails at analysis phase.
- Files: `bazel-multibuild/java/BUILD`, `bazel-multibuild/java/MODULE.bazel`, `bazel-multibuild/README.md`
- Impact: Portfolio claims Bazel expertise; primary demo does not build. Root has no `MODULE.bazel`/`WORKSPACE` — must `cd bazel-multibuild/java` (undocumented).
- Fix approach: Add `rules_oci` / `rules_pkg` / `container_structure_test` deps to `MODULE.bazel` and `load()` statements; or remove container targets until deps are wired.

### Kotlin Detekt rules library has no build system

- Issue: README and CI reference `./gradlew test detekt`, but the subproject has **no** `build.gradle`, `settings.gradle`, or `gradlew`.
- Files: `c0de-quality-and-analysis/kotlin-custom-detekt-rules-library/`, `c0de-quality-and-analysis/kotlin-custom-detekt-rules-library/.github/workflows/ci.yml`
- Impact: Subproject CI workflow cannot succeed; rules cannot be published as a JAR.
- Fix approach: Add Gradle Kotlin DSL project with detekt plugin, wrapper, and `CustomRuleSetProvider` service registration.

### Onboarding CLI missing Gradle repositories

- Issue: `dev-env/onboarding-automation-cli/build.gradle.kts` declares plugins and `application` but no `repositories { mavenCentral() }`.
- Files: `dev-env/onboarding-automation-cli/build.gradle.kts`, `dev-env/onboarding-automation-cli/settings.gradle.kts`
- Impact: `gradle test` fails: "Cannot resolve external dependency org.jetbrains.kotlin:kotlin-stdlib:1.9.24 because no repositories are defined."
- Fix approach: Add `repositories` block; add `gradlew` wrapper; optionally include in root `settings.gradle.kts` composite.

### Empty pyproject.toml stubs in dev-env

- Issue: `dev-env/environment-drift-detector/pyproject.toml` and `dev-env/remote-dev-environment-orchestrator/pyproject.toml` are **empty files** (0 bytes). Source code exists but package is not installable.
- Files: `dev-env/environment-drift-detector/pyproject.toml`, `dev-env/remote-dev-environment-orchestrator/pyproject.toml`, corresponding `src/` trees
- Impact: `pytest` reports "no tests ran"; projects cannot be `pip install -e .`; README usage scripts may fail on imports.
- Fix approach: Fill in `[project]` metadata, setuptools package discovery, dev deps, and add unit tests under `tests/unit/`.

### ForgeX is specification-only

- Issue: `forgex/README.md` is a ~950-line product spec. No `pyproject.toml`, `go.mod`, `package.json`, or source tree.
- Files: `forgex/README.md`
- Impact: Appears in top-level directory listing; duplicates `projgen/` conceptually; cannot be built or demonstrated.
- Fix approach: Mark as "planned" in portfolio index, move to `docs/plans/forgex.md`, or begin minimal CLI scaffold.

### Modular JVM build lacks wrapper

- Issue: `modular-jvm-build/` has Gradle Kotlin DSL (`build.gradle.kts`, multi-module `app`/`api`/`core`) but no `gradlew`. Tests pass with system Gradle 8.10.2 only.
- Files: `modular-jvm-build/build.gradle.kts`, `modular-jvm-build/app/src/test/kotlin/showcase/HealthControllerTest.kt`
- Fix approach: `gradle wrapper` and commit wrapper scripts.

### Meta-documentation drift

- Issue: `PATCH_IMPLEMENTATION_SUMMARY.md` claims SECURITY.md, CONTRIBUTING hygiene, and 10 patches are complete; several claims are stale.
- Files: `PATCH_IMPLEMENTATION_SUMMARY.md`, `SECURITY.md` (empty), `README_PATCHES.md`, `PROJGEN_ENHANCEMENTS.md`
- Impact: Misleading onboarding for contributors; patch workflow scripts (`apply_patches.sh`, `validate_patches.sh`) may be obsolete.
- Fix approach: Archive or update meta-docs; align `SECURITY.md` with claimed content.

---

## Known Bugs

### online-bookstore FastAPI import error

- Symptoms: `pytest` collection error; app cannot start.
- Files: `online-bookstore/src/main.py` (line 1 imports `FastAPI` from `fastapi.responses`), `online-bookstore/test/test_main.py`
- Trigger: `cd online-bookstore && python3 -m pytest`
- Error: `ImportError: cannot import name 'FastAPI' from 'fastapi.responses'`
- Workaround: None; fix import to `from fastapi import FastAPI, HTTPException, Depends, WebSocket` and `from fastapi.responses import JSONResponse`.

### domain-expert-ai tests require editable install

- Symptoms: 6 collection errors from fresh checkout.
- Files: `ai-best-practices-examples/domain-expert-ai/tests/*.py`, `ai-best-practices-examples/domain-expert-ai/pyproject.toml`
- Trigger: `pytest` without `pip install -e ".[dev]"` first
- Error: `ModuleNotFoundError: No module named 'domain_expert_ai'`
- Workaround: `pip install -e ".[dev]"` then 28 tests pass.

### bazel-multibuild BUILD analysis failure

- Symptoms: Build does not start; test targets unreachable.
- Files: `bazel-multibuild/java/BUILD` lines 34–59
- Trigger: `cd bazel-multibuild/java && bazel test //...`
- Error: `name 'oci_image' is not defined`, `name 'tar' is not defined`, `name 'container_structure_test' is not defined`

### otel-demo-stack tests emit OTEL export noise

- Symptoms: Tests pass (3 passed) but stderr floods with connection-refused retries to `127.0.0.1:4318`.
- Files: `otel-demo-stack/api/main.py`, `otel-demo-stack/api/tests/test_api.py`, `otel-demo-stack/collector/otelcol.yaml`
- Trigger: `cd otel-demo-stack/api && pytest` without collector running
- Workaround: Mock/disable OTLP exporter in test config or document `docker compose up` prerequisite.

---

## Security Considerations

### Empty SECURITY.md

- Risk: No documented vulnerability reporting process for a public portfolio.
- Files: `SECURITY.md` (0 bytes), `PATCH_IMPLEMENTATION_SUMMARY.md` (claims policy exists)
- Current mitigation: None in repo.
- Recommendations: Add contact method, supported versions, disclosure timeline per GitHub security best practices.

### Committed virtualenv

- Risk: `projgen/.venv/` contains full site-packages tree in git history; future accidental credential files in venv paths would be committed.
- Files: `projgen/.venv/**` (2,010 tracked files)
- Current mitigation: None.
- Recommendations: Remove from tracking immediately; add `.venv/` to root `.gitignore`; use `pyproject.toml` + `pip install -e .` locally.

### Pre-commit hooks use floating `stable` revs

- Risk: Non-reproducible lint/format behavior across machines and time.
- Files: `.pre-commit-config.yaml` (`black` and `ruff` set to `rev: stable`)
- Recommendations: Pin to specific SHAs or version tags.

---

## Performance Bottlenecks

### Root pytest collection across entire monorepo

- Problem: Collecting all test modules from 40+ trees is slow and error-prone.
- Files: `.github/workflows/ci.yml`, `Makefile`
- Cause: No `testpaths` isolation; duplicate `tests/` package names.
- Improvement path: Matrix per suite or `pytest` invocations scoped to changed paths (already partially done via `dorny/paths-filter` but python job still runs root pytest).

### Clone size from committed venv

- Problem: `projgen/.venv/` alone adds thousands of files to every clone.
- Files: `projgen/.venv/`
- Cause: Missing `.gitignore` entry when venv was committed.
- Improvement path: Remove cached venv; document `python -m venv .venv && pip install -e .` in `projgen/README.md`.

---

## Fragile Areas

### README / portfolio.yaml sync machinery

- Files: `scripts/gen_readme_table.py`, `portfolio.yaml`, `README.md`, `.github/workflows/readme-drift.yml`
- Why fragile: Table is generated from manifest; adding projects requires yaml edit + script run. Currently passes `--check` but only for 5 entries.
- Safe modification: Edit `portfolio.yaml`, run `python3 scripts/gen_readme_table.py`, commit both files.
- Test coverage: Drift check in CI (`readme-drift.yml`) but no test for script itself.

### Legacy design documents path

- Files: `README.md` (Legacy Design Documents section), `src/main/resources/dd-*.md`
- Why fragile: README says "10 of them" but lists 11 bullet items; `dd-advanced-logging-and-tracing.md` is listed but **not present** on disk. Actual count: 10 files.
- Safe modification: Align README list with `ls src/main/resources/`; add missing doc or remove bullet.

### observability vs observability-stack naming

- Files: `PORTFOLIO.md`, `observability/` (docker-compose only), `scripts/create_prs.sh` (references `feat/observability-stack`)
- Why fragile: Competency map points to non-existent `observability-stack` project; actual folder is `observability/`.
- Safe modification: Rename reference in `PORTFOLIO.md` to `observability/` or `otel-demo-stack/`.

### ws-chat-fast listed as stable without tests

- Files: `ws-chat-fast/`, `portfolio.yaml`, `ws-chat-fast/requirements.txt`
- Why fragile: No `tests/` directory; no `pyproject.toml`; only `requirements.txt` with 4 unpinned deps.
- Test coverage: None.

### dev-env scaffold projects with placeholder tests

- Files: `dev-env/remote-dev-environment-orchestrator/tests/unit/.keep`, `dev-env/remote-dev-environment-orchestrator/tests/integration/.keep`, `dev-env/environment-drift-detector/tests/unit/.keep`
- Why fragile: README describes full Temporal workflow and drift detection pipelines; test directories are empty placeholders.
- Test coverage: Zero real tests.

### kotlin-custom-detekt-rules-library CI

- Files: `c0de-quality-and-analysis/kotlin-custom-detekt-rules-library/.github/workflows/ci.yml`
- Why fragile: Workflow assumes Gradle wrapper that does not exist.
- Test coverage: `src/test/kotlin/rules/CustomRulesTest.kt` exists but cannot run without build files.

---

## Scaling Limits

### portfolio.yaml flat list model

- Current capacity: 5 project entries.
- Limit: Adding 40+ subprojects as flat entries will make README table unwieldy.
- Scaling path: Group by suite (`ci-cd-pipelines/*`) with collapsible README sections, or generate multi-table index from nested yaml.

### GitHub Actions path-filter CI

- Current capacity: Single python job on any `**/*.py` change.
- Limit: One broken subproject blocks CI for unrelated Python changes repo-wide.
- Scaling path: Per-suite workflows (many subprojects already have `.github/workflows/ci.yml`) and remove root catch-all pytest.

---

## Dependencies at Risk

### bazel-multibuild MODULE.bazel version skew

- Risk: Warnings that resolved `rules_java` and `rules_jvm_external` versions differ from MODULE.bazel pins.
- Files: `bazel-multibuild/java/MODULE.bazel`, `bazel-multibuild/java/MODULE.bazel.lock`
- Impact: Non-reproducible builds; future Bazel versions may hard-fail.
- Migration plan: Align pinned versions with resolved graph or set `--check_direct_dependencies=off` explicitly with comment.

### Root Gradle JVM toolchain 23

- Risk: Root `build.gradle.kts` sets `jvmToolchain(23)`; many subprojects target Java 21.
- Files: `build.gradle.kts`, `modular-jvm-build/`, `dev-env/onboarding-automation-cli/build.gradle.kts`
- Impact: Contributors without JDK 23 auto-provisioning may fail root build.
- Migration plan: Standardize on Java 21 across portfolio or document toolchain resolver requirements.

### ws-chat-fast unpinned requirements

- Risk: `ws-chat-fast/requirements.txt` lists bare package names without versions.
- Files: `ws-chat-fast/requirements.txt`
- Impact: Non-reproducible installs; demo may break on major FastAPI releases.
- Migration plan: Pin versions or migrate to `pyproject.toml`.

---

## Missing Critical Features

### Portfolio-wide project discovery

- Problem: No single authoritative index of all subprojects.
- Blocks: Recruiters and contributors cannot navigate 6 suites without reading each `README.md`.

### ws-chat-fast test harness

- Problem: Flagship "stable" WebSocket demo has no automated tests.
- Blocks: Regression detection for chat, auth, and metrics endpoints.

### ForgeX implementation

- Problem: Detailed spec with no code.
- Blocks: Cannot demo "polyglot repo generator" distinct from `projgen/`.

### kotlin-custom-detekt-rules-library packaging

- Problem: Rules exist as source but cannot be built or consumed.
- Blocks: Demonstrating custom static analysis plugin contribution competency.

---

## Test Coverage Gaps

### ws-chat-fast

- What's not tested: WebSocket manager, auth, chat routes, templates.
- Files: `ws-chat-fast/app/*.py`
- Risk: Listed **stable** with zero tests.
- Priority: High

### dev-env suite (environment-drift-detector, remote-dev-environment-orchestrator)

- What's not tested: Version comparison, drift classification, Temporal workflow activities.
- Files: `dev-env/environment-drift-detector/src/`, `dev-env/remote-dev-environment-orchestrator/src/`
- Risk: README promises production-like behavior; only `.keep` placeholders in `tests/`.
- Priority: High

### dev-ex/tooling-adoption-tracker

- What's not tested: `make test` target missing; `npm test` passes 1 test when run manually.
- Files: `dev-ex/tooling-adoption-tracker/Makefile`, `dev-ex/tooling-adoption-tracker/package.json`
- Risk: Inconsistent with sibling dev-ex projects that use `make test`.
- Priority: Medium

### online-bookstore

- What's not tested: All endpoints (collection fails before any test runs).
- Files: `online-bookstore/test/test_main.py`, `online-bookstore/src/main.py`
- Risk: **Stable** status is inaccurate.
- Priority: High

### Root CI integration testing

- What's not tested: Monorepo-wide pytest/mypy success.
- Files: `.github/workflows/ci.yml`
- Risk: CI claims test coverage; root invocation fails on 36 modules.
- Priority: High

### bazel-multibuild container targets

- What's not tested: OCI image build, container structure tests.
- Files: `bazel-multibuild/java/BUILD`
- Risk: Container/K8s competency story unsupported by working build.
- Priority: Medium

---

## Unused or Dead Files

### Committed generated and cache artifacts (should not be in git)

| Path | Type | Notes |
|------|------|-------|
| `projgen/.venv/` | Virtualenv | 2,010 tracked files; must be removed |
| `site/` | MkDocs output | 34 tracked HTML/CSS/JS files; regenerate via `mkdocs build` |
| `online-bookstore/src/__pycache__/`, `online-bookstore/test/__pycache__/` | Bytecode | Should be gitignored |
| `otel-demo-stack/api/__pycache__/` | Bytecode | Should be gitignored |
| `platform-audit-template/scripts/__pycache__/` | Bytecode | Should be gitignored |
| `build/` (root) | Gradle output | Partially gitignored; verify not tracked |

### Scaffold-only or spec-only directories

| Path | State |
|------|-------|
| `forgex/README.md` | 950-line spec only; no implementation |
| `dev-env/*/tests/**/.keep` | Empty test placeholders |
| `ai-best-practices-examples/domain-expert-ai/.build/*.stamp` | Build stamp files (untracked in git status) |
| `ai-best-practices-examples/knowledge-qa-system/chroma_data/` | Local vector DB artifacts (untracked) |

### Stale or duplicate documentation

| Path | Issue |
|------|-------|
| `PATCH_IMPLEMENTATION_SUMMARY.md` | Claims completed work that is stale (e.g., SECURITY.md) |
| `README_PATCHES.md` | Patch workflow docs; likely obsolete after manual application |
| `PROJGEN_ENHANCEMENTS.md` | Enhancement notes separate from `projgen/README.md` |
| `PORTFOLIO.md` | References `observability-stack`, missing README anchor |
| `README.md` | Badge placeholders; design doc count mismatch; only 5 projects |
| `rabbit-mq/README.md` | Contains `//TODO: Add a description of the project here` |

### Orphan / auxiliary scripts

| Path | Issue |
|------|-------|
| `apply_patches.sh`, `validate_patches.sh` | Patch application tooling; patches reportedly already applied |
| `scripts/create_prs.sh` | References `feat/observability-stack` branch name that does not match `observability/` dir |

---

## README and Index Gaps

### Root README.md

- Lists **5** projects in auto-generated table (not 6): ws-chat-fast, projgen, rabbit-mq, online-bookstore, bazel-multibuild.
- Badge URLs use `https://github.com/yourusername/portfolio-projects/...` — placeholders (`README.md` lines 3–5).
- Legacy Design Documents section references `src/main/resources/` (path **does** exist) but lists 11 filenames while claiming "10"; `dd-advanced-logging-and-tracing.md` is missing.
- Does not mention: `ai-best-practices-examples/` (5 projects), `ci-cd-pipelines/` (6), `c0de-quality-and-analysis/` (5), `dev-env/` (5), `dev-ex/` (4), `otel-demo-stack/`, `rest-api-test-demo/`, `workflow-api-demo/`, `modular-jvm-build/`, `platform-audit-template/`, `observability/`.

### PORTFOLIO.md

- Competency table has 7 rows; omits AI, code quality, CI/CD pipeline suite, dev-env topics.
- Links to `README.md#portfolio-competency-mapping` — **anchor does not exist** in README.
- References `observability-stack` — directory does not exist (use `observability/` or `otel-demo-stack/`).

### mkdocs.yml

- Nav lists same 5 projects as README; no pages for 6 newer suites.
- Files: `mkdocs.yml`, `docs/index.md`, `docs/design-docs/*/design.md`

### portfolio.yaml

- Source of truth for README table (`scripts/gen_readme_table.py`); only 5 entries.
- `gen_readme_table.py --check` passes — table is in sync but **incomplete** relative to repo contents.

---

## Build Verification Summary (2026-06-22)

| Project / Suite | Command | Result |
|-----------------|---------|--------|
| Root Gradle | `./gradlew test` | PASS (no test sources) |
| rabbit-mq | `./gradlew test` | PASS |
| modular-jvm-build | `gradle test` (no wrapper) | PASS |
| ci-cd-pipelines/* (6) | `make test` per subdir | PASS (9 tests total) |
| dev-ex/* (3 with Makefile) | `make test` | PASS |
| dev-ex/tooling-adoption-tracker | `make test` | FAIL — no test target |
| c0de-quality-and-analysis/* (4 Python) | `pytest` per subdir | PASS |
| kotlin-custom-detekt-rules-library | `./gradlew test` | FAIL — no Gradle project |
| ai-best-practices-examples/knowledge-qa-system | `pytest` | PASS (5) |
| ai-best-practices-examples/chat-ai | `pytest` | PASS (10) |
| ai-best-practices-examples/ai-code-assistant | `pytest` | PASS (38) |
| ai-best-practices-examples/ai-image-video-generator | `pytest` | PASS (16, 1 skipped) |
| ai-best-practices-examples/domain-expert-ai | `pytest` (no install) | FAIL (6 collection errors) |
| ai-best-practices-examples/domain-expert-ai | `pip install -e .[dev] && pytest` | PASS (28) |
| projgen | `pytest` | PASS (9) |
| rest-api-test-demo | `pytest` | PASS (10) |
| workflow-api-demo | `pytest` | PASS (3) |
| otel-demo-stack/api | `pytest` | PASS (3, OTEL export noise) |
| platform-audit-template/scripts | `pytest` | PASS (3) |
| online-bookstore | `pytest` | FAIL (import error) |
| ws-chat-fast | `pytest` | No tests |
| dev-env/environment-drift-detector | `pytest` | No tests |
| dev-env/remote-dev-environment-orchestrator | `pytest` | No tests |
| dev-env/onboarding-automation-cli | `gradle test` | FAIL (no repositories) |
| bazel-multibuild/java | `bazel test //...` | FAIL (BUILD analysis errors) |
| Root monorepo | `pytest -q --continue-on-collection-errors` | 70 passed, **36 errors** |
| Root monorepo | `ruff check .` | 6 fixable errors |
| Root monorepo | `mypy .` | FAIL (duplicate conftest) |

---

*Concerns audit: 2026-06-22*
