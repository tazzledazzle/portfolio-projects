# API Breaking-Change Detector

## Overview
This project compares OpenAPI specs from the current branch and `main`, classifies diffs by compatibility rules, and blocks merges on unintended breaking changes.

## Architecture
- `src/spec`: OpenAPI loading and normalization.
- `src/diff`: endpoint/schema diff engine.
- `src/classify`: breaking vs non-breaking rule classifier.
- `src/ci`: exit-code and report adapters for pipelines.
- `config`: intentional break allowlist and strictness settings.

## Use Cases
- Enforce backward-compatible API evolution.
- Surface high-signal API risk in pull requests.
- Allow intentional breaks with explicit approvals.

## Usage
1. Export baseline spec from `main`.
2. Generate branch spec.
3. Run detector with both specs and policy config.
4. Fail CI when unapproved breaking changes are found.

## Control Flow
1. Parse baseline and candidate OpenAPI documents.
2. Compute path, operation, and schema diffs.
3. Classify each diff using compatibility rules.
4. Apply allowlist overrides for approved breakages.
5. Emit summary report and exit code for merge gating.

## Project Structure
```text
api-breaking-change-detector/
  .github/workflows/ci.yml
  config/compatibility-policy.yml
  scripts/check-api-compat.sh
  src/spec/loader.py
  src/diff/openapi_diff.py
  src/classify/compatibility_rules.py
  src/ci/merge_gate.py
  tests/test_compatibility_rules.py
```
