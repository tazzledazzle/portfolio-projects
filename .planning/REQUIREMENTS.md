# Requirements: Portfolio Projects — P0/P1 Fix

**Defined:** 2026-07-09
**Core Value:** A credible, runnable, and well-indexed portfolio monorepo that accurately represents engineering competency.

## v1 Requirements

Requirements for the P0/P1 remediation pass. Maps to Phase 1.

### P0 — Critical Bugs (must pass before portfolio is usable)

- [ ] **P0-01**: `projgen/.venv/` (2,010 files), all `__pycache__/` dirs, and `.pytest_cache/` dirs are removed from git tracking and covered by root `.gitignore` entries
- [ ] **P0-02**: `online-bookstore/src/main.py` imports `FastAPI` correctly from `fastapi` (not `fastapi.responses`); `cd online-bookstore && pytest` collects and passes all tests
- [ ] **P0-03**: `bazel-multibuild/java/BUILD` loads `oci_image`, `tar`, and `container_structure_test` rules from `MODULE.bazel`; `cd bazel-multibuild/java && bazel test //...` reaches at least the test execution phase (no analysis errors)

### P1 — High-Priority Gaps

- [ ] **P1-01**: `portfolio.yaml` contains entries for all major project suites (ai-best-practices-examples, c0de-quality-and-analysis, ci-cd-pipelines, dev-env, dev-ex, and standalone demos); `scripts/gen_readme_table.py --check` passes; README table reflects the expanded set
- [ ] **P1-02**: README.md badge URLs (lines 3–5) contain the real GitHub username/org — no `yourusername` placeholder appears in any badge URL
- [ ] **P1-03**: Root CI pytest strategy changed so that `pytest` from repo root produces zero collection errors; CI workflow passes on any Python file change
- [ ] **P1-04**: Root CI mypy strategy changed so `mypy` from repo root does not fail on duplicate `conftest` module names; per-package scoping or `--explicit-package-bases` applied
- [ ] **P1-05**: `dev-env/onboarding-automation-cli/build.gradle.kts` has a `repositories { mavenCentral() }` block and a `gradlew` wrapper is committed; `./gradlew test` resolves all dependencies and runs
- [ ] **P1-06**: `c0de-quality-and-analysis/kotlin-custom-detekt-rules-library/` has a working Gradle Kotlin DSL project (`build.gradle.kts`, `settings.gradle.kts`, `gradlew`); `./gradlew test detekt` compiles and runs the custom rule tests

## Out of Scope

| Feature | Reason |
|---------|--------|
| P2/P3 issues | Tracked as separate beads; addressed after P0/P1 pass |
| New feature development | Phase 1 is remediation-only |
| MkDocs nav expansion | Dependent on P1-01 portfolio.yaml work; separate phase |

## Traceability

| Requirement | Phase | Beads Issue | Status |
|-------------|-------|-------------|--------|
| P0-01 | Phase 1 | portfolio-projects-8ax | Pending |
| P0-02 | Phase 1 | portfolio-projects-tye | Pending |
| P0-03 | Phase 1 | portfolio-projects-96h | Pending |
| P1-01 | Phase 1 | portfolio-projects-b7p | Pending |
| P1-02 | Phase 1 | portfolio-projects-ag8 | Pending |
| P1-03 | Phase 1 | portfolio-projects-oir | Pending |
| P1-04 | Phase 1 | portfolio-projects-840 | Pending |
| P1-05 | Phase 1 | portfolio-projects-ded | Pending |
| P1-06 | Phase 1 | portfolio-projects-3xf | Pending |
