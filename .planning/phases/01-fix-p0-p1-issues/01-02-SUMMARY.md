---
phase: 01-fix-p0-p1-issues
plan: "02"
subsystem: online-bookstore
tags: [python, fastapi, pytest, dependencies]
dependency_graph:
  requires: []
  provides: [online-bookstore-tests-passing]
  affects: [online-bookstore/src/requirements.txt, online-bookstore/pyproject.toml]
tech_stack:
  added: [python-multipart, httpx]
  patterns: [fastapi-oauth2, testclient]
key_files:
  modified:
    - online-bookstore/src/requirements.txt
    - online-bookstore/pyproject.toml
decisions:
  - Added python-multipart and httpx unpinned to match existing requirements.txt style
  - Used setuptools build backend in pyproject.toml matching project's flat src/ layout
metrics:
  duration: "~3 minutes"
  completed: "2026-07-09"
---

# Phase 01 Plan 02: Fix online-bookstore pytest — missing python-multipart and pyproject.toml Summary

**One-liner:** Added python-multipart and httpx to requirements, wrote valid pyproject.toml so `cd online-bookstore && python3 -m pytest test/ -q` exits 0 with 6 passed.

## What Was Built

Fixed two root causes preventing `cd online-bookstore && python3 -m pytest test/ -q` from succeeding:

1. **Missing python-multipart** — FastAPI raises `RuntimeError` at import time when `OAuth2PasswordRequestForm` is used but `python-multipart` is not installed. Added `python-multipart` and `httpx` (needed by `fastapi.testclient`) to `online-bookstore/src/requirements.txt`.

2. **Empty pyproject.toml** — The portfolio CI script calls `pip install -e .` for directories with a `pyproject.toml`. An empty file caused a silent install failure. Wrote a valid `[project]` table with name, version, requires-python, runtime dependencies, and a setuptools build-system block.

## Task Results

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Add python-multipart to requirements | 9931710 | online-bookstore/src/requirements.txt |
| 2 | Fill pyproject.toml and verify pytest passes | 05315ac | online-bookstore/pyproject.toml |

## Verification

```
cd online-bookstore && python3 -m pytest test/ -q
6 passed in 0.11s
```

All acceptance criteria met:
- `grep "python-multipart" online-bookstore/src/requirements.txt` — matches
- `grep "httpx" online-bookstore/src/requirements.txt` — matches
- `wc -c < online-bookstore/pyproject.toml` — 281 bytes (non-empty)
- `grep "\[project\]" online-bookstore/pyproject.toml` — matches
- pytest exit code 0, final line: "6 passed in 0.11s"

## Deviations from Plan

None — plan executed exactly as written.

## Known Stubs

None.

## Threat Flags

None — no new network endpoints, auth paths, or trust boundary changes introduced. Only dependency metadata and package metadata files modified.

## Self-Check: PASSED

- online-bookstore/src/requirements.txt: FOUND
- online-bookstore/pyproject.toml: FOUND (281 bytes)
- Task 1 commit 9931710: FOUND
- Task 2 commit 05315ac: FOUND
- pytest: 6 passed, exit 0
