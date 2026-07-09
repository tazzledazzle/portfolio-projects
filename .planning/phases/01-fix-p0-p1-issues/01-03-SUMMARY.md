---
phase: 01-fix-p0-p1-issues
plan: "03"
subsystem: bazel-multibuild/java
tags: [bazel, maven, security, supply-chain]
dependency_graph:
  requires: []
  provides: [bazel-multibuild/java/MODULE.bazel clean Maven repos]
  affects: [bazel-multibuild/java build]
tech_stack:
  added: []
  patterns: [Bazel module extensions, rules_jvm_external, Maven Central only]
key_files:
  modified:
    - bazel-multibuild/java/MODULE.bazel
decisions:
  - "Use only https://repo.maven.apache.org/maven2 as Maven repository; jcenter and plain-HTTP uk.maven.org removed"
metrics:
  duration: "~3 minutes"
  completed: "2026-07-09"
  tasks_completed: 1
  tasks_total: 1
requirements_satisfied: [P0-03]
---

# Phase 01 Plan 03: Remove Deprecated Maven Repositories Summary

**One-liner:** Removed jcenter.bintray.com (shutdown/supply-chain risk) and plain-HTTP Maven URL from MODULE.bazel, leaving only HTTPS Maven Central; bazel test //... passes with 1 test.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Verify bazel test passes and remove deprecated jcenter repository | 35a9d66 | bazel-multibuild/java/MODULE.bazel |

## What Was Done

**Task 1:** Ran initial `bazel test //...` to confirm baseline (passed). Edited `bazel-multibuild/java/MODULE.bazel` to remove two insecure Maven repository URLs from the `maven.install` repositories list:
- `https://jcenter.bintray.com/` — shut down, could be re-registered by a malicious actor (T-03-01)
- `http://uk.maven.org/maven2` — plain HTTP, susceptible to MITM attacks (T-03-02)

Only `https://repo.maven.apache.org/maven2` remains. Re-ran `bazel test //...` to confirm no regression: `//:myproject_test PASSED`.

## Verification Results

```
bazel test //...  ->  Build completed successfully
                      //:myproject_test  PASSED in 0.5s
                      1 test passes.

grep "jcenter.bintray.com" MODULE.bazel  ->  exit 1 (no matches)
grep "http://" MODULE.bazel              ->  exit 1 (no matches)
```

## Deviations from Plan

None - plan executed exactly as written.

## Known Stubs

None.

## Threat Surface Scan

No new security surface introduced. Two threat mitigations applied as planned:
- T-03-01 mitigated: jcenter URL removed
- T-03-02 mitigated: plain-HTTP URL removed

## Self-Check: PASSED

- [x] `bazel-multibuild/java/MODULE.bazel` modified and committed at 35a9d66
- [x] `bazel test //...` exits 0 with `//:myproject_test PASSED`
- [x] No jcenter.bintray.com in MODULE.bazel
- [x] No plain HTTP URLs in MODULE.bazel
- [x] BUILD file unchanged (container targets remain commented out)
