---
phase: 01-fix-p0-p1-issues
plan: "04"
subsystem: test-infrastructure
tags:
  - pytest
  - monorepo
  - test-isolation
  - makefile
dependency_graph:
  requires: []
  provides:
    - root-pytest-safe-collection
    - makefile-test-uses-per-suite-runner
  affects:
    - developer-test-workflow
    - ci-cd-pipelines
tech_stack:
  added: []
  patterns:
    - pytest.ini norecursedirs guard
    - per-suite test runner via Makefile
key_files:
  created:
    - pytest.ini
  modified:
    - Makefile
decisions:
  - "Used norecursedirs = * instead of empty testpaths alone — pytest 9.x falls back to CWD when testpaths points to a non-existent path; norecursedirs = * guarantees zero collection from root"
  - "P1-01 already resolved: gen_readme_table.py --check exits 0 (portfolio.yaml complete)"
  - "P1-02 already resolved: README.md contains no yourusername placeholders"
metrics:
  duration: "~80s"
  completed: "2026-07-09T18:32:43Z"
  tasks_completed: 2
  files_changed: 2
---

# Phase 01 Plan 04: Pytest Root Guard and Makefile Test Target Summary

Root pytest.ini added with `norecursedirs = *` to prevent 64 collection errors when running bare `pytest` from the monorepo root; Makefile `test` target updated to call `scripts/run_portfolio_pytest.sh` instead of root-level pytest.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Add root pytest.ini to prevent accidental monorepo-wide collection | 3b6d415 | pytest.ini |
| 2 | Update Makefile test target to use run_portfolio_pytest.sh | 0bd66b7 | Makefile |

## Changes Made

### Task 1: pytest.ini (commit 3b6d415)

Created `pytest.ini` at repo root with:

```ini
[pytest]
testpaths =
norecursedirs = *
addopts = --tb=short
```

**Verification results:**
- `python3 -m pytest --collect-only -q 2>&1 | grep -c "error"` = 0
- `python3 -m pytest --collect-only -q 2>&1 | grep -c "ImportPathMismatch"` = 0
- `python3 -m pytest --collect-only -q` output = "no tests collected in 0.00s"

### Task 2: Makefile (commit 0bd66b7)

Replaced the pytest invocation in the `test` target:

**Before:**
```makefile
@if command -v pytest >/dev/null 2>&1; then \
    pytest -v --tb=short --continue-on-collection-errors || echo "Some Python tests failed"; \
else \
    echo "pytest not installed. Install with: pip install pytest"; \
fi
```

**After:**
```makefile
bash scripts/run_portfolio_pytest.sh
```

`make -n test` exits 0 (no syntax errors).

## Already-Resolved Issues (Documented Here)

### P1-01: portfolio.yaml completeness
- `python3 scripts/gen_readme_table.py --check` exits 0
- portfolio.yaml has 297 lines covering all 6 portfolio suites (37 entries)
- No changes needed

### P1-02: README.md badge placeholders
- `grep "yourusername" README.md` returns no matches
- All badges already reference `tazzledazzle`
- No changes needed

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Added norecursedirs = * to pytest.ini**
- **Found during:** Task 1 verification
- **Issue:** The plan stated that `testpaths =` (empty) would prevent collection, but pytest 9.x falls back to recursive CWD scanning with a PytestConfigWarning when testpaths is empty or points to a non-existent path. The 64 collection errors persisted.
- **Fix:** Added `norecursedirs = *` which prevents pytest from descending into any subdirectory when invoked from the repo root. This produces "no tests collected in 0.00s" with zero errors and exit code 5.
- **Files modified:** pytest.ini
- **Commit:** 3b6d415

## Threat Surface Scan

No new network endpoints, auth paths, file access patterns, or schema changes introduced. Changes are limited to developer tooling configuration files (pytest.ini, Makefile).

## Known Stubs

None — no data flowing to UI rendering or placeholder text.

## Self-Check

- [x] pytest.ini exists at repo root: FOUND
- [x] Makefile contains run_portfolio_pytest.sh: FOUND
- [x] Old pytest invocation removed from Makefile: CONFIRMED
- [x] Commit 3b6d415 exists: CONFIRMED
- [x] Commit 0bd66b7 exists: CONFIRMED
- [x] `python3 -m pytest --collect-only -q` produces 0 error lines: CONFIRMED
- [x] `make -n test` exits 0: CONFIRMED (returncode 0 via subprocess)

## Self-Check: PASSED
