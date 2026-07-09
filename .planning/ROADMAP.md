# Roadmap: Portfolio Projects Remediation

## Overview

This roadmap covers the remediation of critical (P0) and high-priority (P1) issues identified in the portfolio monorepo codebase scan. The goal is to produce a portfolio that builds cleanly, tests reliably, and accurately represents all 48+ subprojects to recruiters and collaborators.

## Phases

- [ ] **Phase 1: Fix P0 and P1 Portfolio Issues** - Remove committed artifacts, fix broken imports and builds, repair CI strategy, expand portfolio index, fix placeholder content

## Phase Details

### Phase 1: Fix P0 and P1 Portfolio Issues

**Goal**: All P0 and P1 issues resolved — repo builds cleanly, CI passes, portfolio index is complete, and no placeholder content remains
**Depends on**: Nothing (first phase)
**Requirements**: P0-01, P0-02, P0-03, P1-01, P1-02, P1-03, P1-04, P1-05, P1-06
**Success Criteria** (what must be TRUE):

  1. `git ls-files projgen/.venv` returns empty; root `.gitignore` covers Python virtualenvs and caches
  2. `cd online-bookstore && pytest` passes all tests with no collection errors
  3. `cd bazel-multibuild/java && bazel test //...` reaches test execution phase (no BUILD analysis errors)
  4. `portfolio.yaml` lists entries for all 6 major suites; `scripts/gen_readme_table.py --check` passes
  5. `grep -r "yourusername" README.md` returns no results
  6. Root pytest produces zero collection errors (matrix strategy or isolated testpaths)
  7. Root mypy completes without duplicate-conftest failure
  8. `cd dev-env/onboarding-automation-cli && ./gradlew test` resolves all dependencies and passes
  9. `cd c0de-quality-and-analysis/kotlin-custom-detekt-rules-library && ./gradlew test detekt` compiles and passes

**Plans**: 7 plans
Plans:

- [ ] 01-01-PLAN.md — Remove tracked __pycache__ and bytecode artifacts via git rm --cached
- [ ] 01-02-PLAN.md — Fix online-bookstore pytest (add python-multipart, fill pyproject.toml)
- [ ] 01-03-PLAN.md — Verify bazel test passes; remove deprecated jcenter and HTTP Maven URL from MODULE.bazel
- [ ] 01-04-PLAN.md — Add pytest.ini to prevent root collection errors; fix Makefile test target; verify P1-01 and P1-02 already resolved
- [ ] 01-05-PLAN.md — Add monorepo-safe mypy step to CI python job (--explicit-package-bases)
- [ ] 01-06-PLAN.md — Fix onboarding-automation-cli Gradle build (settings.gradle.kts + Kotlin plugin)
- [ ] 01-07-PLAN.md — Fix kotlin-custom-detekt-rules-library (add detekt plugin version)
