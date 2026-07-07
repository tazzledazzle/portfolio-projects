# License Compliance Scanner

## Overview
This project scans transitive dependencies, resolves license metadata using SPDX identifiers, and blocks CI when policy-incompatible licenses are introduced.

## Architecture
- `src/ingest`: dependency tree readers (Gradle-first, extensible).
- `src/matching`: SPDX normalization and lookup.
- `src/policy`: allow/deny and exception evaluation.
- `src/output`: SBOM generation and CI-friendly summaries.
- `config`: policy and exception definitions.

## Use Cases
- Prevent accidental adoption of restricted licenses.
- Produce auditable SBOM artifacts for compliance reviews.
- Enforce policy as a merge gate in CI.

## Usage
1. Export dependency graph from build tooling.
2. Run scanner with `config/license-policy.yml`.
3. Review report and generated SBOM.
4. Fail pipeline if violations are detected.

## Control Flow
1. Ingest dependency graph.
2. Resolve package licenses via SPDX data.
3. Evaluate each dependency against policy rules.
4. Emit violations and write SBOM artifact.
5. Return non-zero exit code on policy failure.

## Project Structure
```text
license-compliance-scanner/
  .github/workflows/ci.yml
  config/license-policy.yml
  scripts/run-scan.sh
  src/ingest/gradle_tree_parser.py
  src/matching/spdx_matcher.py
  src/policy/policy_engine.py
  src/output/sbom_writer.py
  tests/test_policy_engine.py
```
